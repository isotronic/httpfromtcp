package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

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

func handleRequest(w io.Writer, req *request.Request) *server.HandlerError {
	handlerError := server.HandlerError{}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		handlerError.StatusCode = response.StatusBadRequest
		handlerError.Message = "Your problem is not my problem\n"
		return &handlerError
	}
	
	if req.RequestLine.RequestTarget == "/myproblem" {
		handlerError.StatusCode = response.StatusInternalServerError
		handlerError.Message = "Woopsie, my bad\n"
		return &handlerError
	}

	w.Write([]byte("All good, frfr\n"))
	return nil
}