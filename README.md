
# Octave MCP Server

A Model Context Protocol (MCP) server for executing Octave scripts non-interactively.

![Octave MCP logo](https://github.com/fmcato/octave-mcp/raw/main/assets/logo-200.png)

## Features

- Execute Octave scripts via MCP protocol
- Supports both HTTP and stdio communication modes
- Built-in security for HTTP mode (localhost only)
- Automatic Octave installation verification

## Prerequisites

- Go 1.21+ (for building/running)
- GNU Octave installed and available in PATH (tested version 8.4.0)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/fmcato/octave-mcp.git
   cd octave-mcp
   ```

2. Build the server:
   ```bash
   go build ./cmd/octave-server
   ```

## Usage

### HTTP Mode

Start the HTTP server:
```bash
./octave-server -http localhost:8080
```

### Stdio Mode

Start the stdio server:
```bash
./octave-server
```

### MCP Tool Usage

The server provides two tools:

1. `run_octave` - Execute Octave scripts:
```json
{
  "script": "string"
}
```

Example:
```json
{
  "script": "disp('Hello from Octave');"
}
```

2. `generate_plot` - Generate plots from Octave scripts:
```json
{
  "script": "string",
  "format": "png|svg"
}
```

Example:
```json
{
  "script": "plot([1,2,3,4]);",
  "format": "png"
}
```

**Plot Generation Notes:**
- Requires Octave's `qt` graphics toolkit
- Output formats supported: PNG or SVG
- Maximum output size: 1MB

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
