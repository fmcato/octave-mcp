# Use Go 1.24 as the base image
FROM golang:1.24

# Install Octave and dependencies
RUN apt-get update && apt-get install -y \
    gnuplot \
    octave \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o octave-server ./cmd/octave-server

# Expose port
EXPOSE 8080

# Run the application in HTTP mode by default
CMD ["./octave-server", "-http", "0.0.0.0:8080"]