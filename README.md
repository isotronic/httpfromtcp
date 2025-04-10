# HTTP from TCP - A TCP-based HTTP/1.1 Server Implementation in Go

My semi-guided attempt at building an HTTP/1.1 server implementation in Go, built directly on top of TCP. This project demonstrates how to handle HTTP requests at the TCP level, implementing some of the HTTP/1.1 protocol specifications. It also includes a TCP listener and UDP sender for debugging and testing purposes.

## Features

- Pure TCP-based HTTP/1.1 server implementation
- Support for common HTTP methods (GET, POST)
- Chunked transfer encoding support
- HTTP header parsing and manipulation
- Request body handling with Content-Length validation
- Multiple connection handling with goroutines
- Custom response writer with status codes, headers, and body support
- Example handlers for different HTTP scenarios

## Getting Started

### Prerequisites

- Go 1.24.2 or higher

### Installation

```bash
git clone https://github.com/isotronic/httpfromtcp.git
cd httpfromtcp
go mod download
```

### Running the Servers

The project includes several example servers:

1. HTTP Server (main implementation):

```bash
go run cmd/httpserver/main.go
```

This starts an HTTP server on port 42069 with the following endpoints:

- `/` - Returns a 200 OK response with HTML content
- `/video` - Serves a video file (requires `assets/vim.mp4`)
- `/yourproblem` - Returns a 400 Bad Request response
- `/myproblem` - Returns a 500 Internal Server Error response
- `/httpbin/*` - Proxies requests to httpbin.org with chunked transfer encoding

2. TCP Listener (for debugging):

```bash
go run cmd/tcplistener/main.go
```

This starts a basic TCP server that prints received HTTP requests.

3. UDP Sender (for testing):

```bash
go run cmd/udpsender/main.go
```

A simple UDP client for testing purposes.

## Project Structure

```
.
├── cmd/
│   ├── httpserver/    # Main HTTP server implementation
│   ├── tcplistener/   # TCP debugging server
│   └── udpsender/     # UDP test client
├── internal/
│   ├── headers/       # HTTP headers implementation
│   ├── request/       # HTTP request parsing
│   ├── response/      # HTTP response writing
│   └── server/        # Server core functionality
└── assets/           # Static assets (not included in repo)
```

## Testing

Run the test suite:

```bash
go test ./...
```

The project includes comprehensive tests for request parsing, header handling, and server functionality.

## Implementation Details

- Built directly on top of TCP connections using Go's `net` package
- Implements HTTP/1.1 protocol features including:
  - Request line parsing
  - Header parsing and validation
  - Body handling with Content-Length
  - Chunked transfer encoding
  - Trailer headers
- Uses a state machine approach for parsing HTTP requests
- Supports concurrent connections using goroutines

## License

This project is licensed under the MIT License - see the LICENSE file for details.
