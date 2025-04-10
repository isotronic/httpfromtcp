package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/isotronic/httpfromtcp/internal/headers"
	"github.com/isotronic/httpfromtcp/internal/request"
	"github.com/isotronic/httpfromtcp/internal/response"
	"github.com/isotronic/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handleRequest)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handleRequest(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		url := "https://httpbin.org/" + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

		res, err := http.Get(url)
		if err != nil {
			h := response.GetDefaultHeaders(len(err.Error()))
			w.WriteStatusLine(response.StatusInternalServerError)
			w.WriteHeaders(h)
			w.WriteBody([]byte(err.Error()))
			return
		}
		defer res.Body.Close()

		h := response.GetDefaultHeaders(0)
		h.Remove("Content-Length")
		h.Add("Transfer-Encoding", "chunked")
		h.Add("Trailer", "X-Content-SHA256, X-Content-Length")
		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(h)

		rawBody := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := res.Body.Read(buf)
			if n > 0 {
				rawBody = append(rawBody, buf[:n]...)
				w.WriteChunkedBody(buf[:n])
			}

			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err)
				break
			}
		}
		t := headers.NewHeaders()
		sum := sha256.Sum256(rawBody)
		length := strconv.Itoa(len(rawBody))
		t.Add("X-Content-SHA256", fmt.Sprintf("%x", sum))
		t.Add("X-Content-Length", length)
		w.WriteChunkedBodyDone()
		w.WriteTrailers(t)
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		f, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Println(err)
		}
		h := response.GetDefaultHeaders(len(f))
		h.Override("Content-Type", "video/mp4")
		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(h)
		w.WriteBody(f)
		return
	}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		body := `
			<html>
				<head>
					<title>400 Bad Request</title>
				</head>
				<body>
					<h1>Bad Request</h1>
					<p>Your request honestly kinda sucked.</p>
				</body>
			</html>
		`
		h := response.GetDefaultHeaders(len(body))
		h.Override("Content-Type", "text/html")
		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(h)
		w.WriteBody([]byte(body))
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		body := `
			<html>
				<head>
					<title>500 Internal Server Error</title>
				</head>
				<body>
					<h1>Internal Server Error</h1>
					<p>Okay, you know what? This one is on me.</p>
				</body>
			</html>
		`
		h := response.GetDefaultHeaders(len(body))
		h.Override("Content-Type", "text/html")
		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(h)
		w.WriteBody([]byte(body))
		return
	}

	body := `
		<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			</body>
		</html>
	`
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}
