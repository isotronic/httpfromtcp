package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
		conn.Close()
		log.Println("connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	var currentLine string
	c := make(chan string)
	go func() {
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err == io.EOF {
				if currentLine != "" {
					c <- currentLine
				}
				break
			}

			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := range parts[:len(parts)-1] {
				c <- currentLine + parts[i]
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
		close(c)
	}()
	return c
}
