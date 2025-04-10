package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/isotronic/httpfromtcp/internal/headers"
	"github.com/isotronic/httpfromtcp/internal/request"
	"github.com/isotronic/httpfromtcp/internal/response"
)

func handleRequest(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handleChunk(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		handleVideo(w)
		return
	}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		handle400(w)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		handle500(w)
		return
	}

	if req.RequestLine.RequestTarget == "/" {
		handle200(w)
		return
	}
}

func handleChunk(w *response.Writer, req *request.Request) {
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
}

func handleVideo(w *response.Writer) {
	f, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Println(err)
	}
	h := response.GetDefaultHeaders(len(f))
	h.Override("Content-Type", "video/mp4")
	w.WriteStatusLine(response.StatusOK)
	w.WriteHeaders(h)
	w.WriteBody(f)
}

func handle500(w *response.Writer) {
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
}

func handle400(w *response.Writer) {
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
}

func handle200(w *response.Writer) {
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