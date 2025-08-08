package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fmcato/octave-mcp/octave"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RunOctaveParams struct {
	Script string `json:"script"`
}

type Server struct {
	mcpServer *mcp.Server
	runner    *octave.Runner
}

func New() *Server {
	return &Server{
		runner: octave.NewRunner(),
		mcpServer: mcp.NewServer(&mcp.Implementation{
			Name:    "octave-mcp",
			Version: "1.0.0",
		}, nil),
	}
}

func (s *Server) RegisterHandlers() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "run_octave",
		Description: "Executes GNU Octave scripts non-interactively. Ideal for off-loading calculations from the LLM.",
	}, s.runOctaveHandler)
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
