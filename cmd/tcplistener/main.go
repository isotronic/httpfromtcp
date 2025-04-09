package main

import (
	"fmt"
	"log"
	"net"

	"github.com/isotronic/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v\n", err)
			continue
		}

		log.Println("connection accepted")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("error parsing request: %v", err)
			continue
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", req.RequestLine.Method)
		fmt.Printf("- Target: %v\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for header := range req.Headers {
			fmt.Printf("- %v: %v\n", header, req.Headers[header])
		}

		fmt.Println("Body:")
		fmt.Printf("%v\n", string(req.Body))

		conn.Close()
		log.Println("connection closed")
	}
}
