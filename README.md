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
- Output formats supported: PNG or SVG


## Running with Docker

You can run the Octave MCP server using Docker for easier deployment and isolation.

The Docker image uses a multi-stage build process:
- Build stage: Uses `golang:alpine` for compiling the binary
- Runtime stage: Uses the official `gnuoctave/octave` image
- This approach reduces the final image size and improves security

### Using Docker Compose (Recommended)

Build and run the server:
```bash
docker-compose up --build
```

The server will be available at `http://localhost:8080`.

### Using Docker Directly

Build the image:
```bash
docker build -t octave-mcp -f docker/Dockerfile .
```

Run the container:
```bash
docker run -p 8080:8080 octave-mcp
```

### Testing the Docker Image

To verify the Docker image is working correctly:

1. Start the container:
   ```bash
   docker-compose up -d
   ```

2. Use a tool like MCP Inspector to verify the server is accessible at http://localhost:8080/mcp

3. Check the logs:
   ```bash
   docker-compose logs octave-mcp
   ```

4. Stop the container:
   ```bash
   docker-compose down
   ```

## Configuration

The server accepts the following flags:
- `-http`: HTTP address to listen on (empty for stdio mode)

## Environment Variables

The following environment variables can be used to configure server behavior:

- `OCTAVE_SCRIPT_TIMEOUT`: Script execution timeout in seconds (default: 10)
- `OCTAVE_CONCURRENCY_LIMIT`: Maximum concurrent executions (default: 10)
- `OCTAVE_SCRIPT_LENGTH_LIMIT`: Maximum script length in characters (default: 10000)
- `OCTAVE_MCP_ALLOW_NON_LOCALHOST`: Set to `true` to allow non-localhost connections (default: `false`). Use with caution in production environments.

## Security

- Scans scripts for dangerous patterns
- Filters output to remove sensitive information
- Uses temporary directories with restricted permissions

When running in HTTP mode:
- Only accepts connections from localhost (unless `OCTAVE_MCP_ALLOW_NON_LOCALHOST=true` is set)
- Implements strict CORS and security headers
- Validates request origins

## License

This project is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for the full license text.
