# Build stage: Use golang:alpine for smaller footprint
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application as a static binary for Alpine compatibility
RUN CGO_ENABLED=0 GOOS=linux go build -o octave-server ./cmd/octave-server

# Runtime stage: Use the official gnuoctave/octave image
FROM gnuoctave/octave:10.1.0

# Set working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/octave-server .

# Expose port
EXPOSE 8080

# Run the application in HTTP mode by default
CMD ["./octave-server", "-http", "0.0.0.0:8080"]
