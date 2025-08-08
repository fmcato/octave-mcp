# Octave MCP Server

A Model Context Protocol (MCP) server for executing Octave scripts non-interactively.

## Features

- Execute Octave scripts via MCP protocol
- Supports both HTTP and stdio communication modes
- Built-in security for HTTP mode (localhost only)
- Automatic Octave installation verification

## Prerequisites

- Go 1.20+ (for building/running)
- Octave installed and available in PATH

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/fmcato/octave-mcp.git
   cd octave-mcp
   ```

2. Build the server:
   ```bash
   go build ./cmd/main.go
   ```

## Usage

### HTTP Mode

Start the HTTP server:
```bash
./main -http localhost:8080
```

### Stdio Mode

Start the stdio server:
```bash
./main
```

### MCP Tool Usage

The server provides a `run_octave` tool with the following schema:
```json
{
  "script": "string"
}
```

Example tool call:
```json
{
  "script": "disp('Hello from Octave');"
}
```

## Configuration

The server accepts the following flags:
- `-http`: HTTP address to listen on (empty for stdio mode)

## Security

When running in HTTP mode:
- Only accepts connections from localhost
- Implements strict CORS and security headers
- Validates request origins

## License

This project is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for the full license text.
