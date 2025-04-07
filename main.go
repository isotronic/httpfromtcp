package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	c := getLinesChannel(file)
	for line := range c {
		fmt.Printf("read: %s\n", line)
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
