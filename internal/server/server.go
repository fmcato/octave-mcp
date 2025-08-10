package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fmcato/octave-mcp/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RunOctaveParams struct {
	Script string `json:"script"`
}

type GeneratePlotParams struct {
	Script string `json:"script"`
	Format string `json:"format"` // "png" or "svg"
}

type Server struct {
	mcpServer *mcp.Server
	runner    *domain.Runner
}

func New() *Server {
	return &Server{
		runner: domain.NewRunner(),
		mcpServer: mcp.NewServer(&mcp.Implementation{
			Name:    "octave-mcp",
			Version: "1.0.0",
		}, nil),
	}
}

func (s *Server) RegisterHandlers() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "run_octave",
		Description: "Executes a GNU Octave script non-interactively. Ideal for off-loading calculations from the LLM.",
	}, s.runOctaveHandler)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "generate_plot",
		Description: "Generate a plot from a GNU Octave script. Returns image data in specified format (png/svg). Use the plot() command and any other one for labels, legend, etc. Do not try to set graphics toolkit or other format options.",
	}, s.generatePlotHandler)
}

func (s *Server) RunHTTP(addr string) error {
	if !strings.Contains(addr, "localhost") && !strings.Contains(addr, "127.0.0.1") {
		return fmt.Errorf("HTTP server must bind to localhost for security")
	}

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, &mcp.StreamableHTTPOptions{})

	log.Printf("Starting HTTP server on %s", addr)
	http.Handle("/mcp", securityMiddleware(handler))
	return http.ListenAndServe(addr, nil)
}

func (s *Server) RunStdio() error {
	log.Println("Starting stdio server")
	transport := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)
	return s.mcpServer.Run(context.Background(), transport)
}

func (s *Server) runOctaveHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[RunOctaveParams]) (*mcp.CallToolResultFor[any], error) {
	if params.Arguments.Script == "" {
		return nil, fmt.Errorf("script parameter is required")
	}

	result, err := s.runner.ExecuteScript(ctx, params.Arguments.Script)

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		IsError: false,
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

func (s *Server) generatePlotHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[GeneratePlotParams]) (*mcp.CallToolResultFor[any], error) {
	if params.Arguments.Script == "" {
		return nil, fmt.Errorf("script parameter is required")
	}

	imgData, err := s.runner.GeneratePlot(ctx, params.Arguments.Script, params.Arguments.Format)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		}, nil
	}

	var mimeType string
	switch params.Arguments.Format {
	case "svg":
		mimeType = "image/svg+xml"
	case "png":
		mimeType = "image/png"
	default:
		mimeType = "application/octet-stream"
	}

	return &mcp.CallToolResultFor[any]{
		IsError: false,
		Content: []mcp.Content{&mcp.ImageContent{Data: imgData, MIMEType: mimeType}},
	}, nil
}

func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && !strings.HasPrefix(origin, "http://localhost") {
			http.Error(w, "Invalid origin", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		next.ServeHTTP(w, r)
	})
}
