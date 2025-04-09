package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/isotronic/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       RequestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type RequestState int

const (
	initialized RequestState = iota
	parsingHeaders
	done
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := Request{
		State:   initialized,
		Headers: headers.Headers{},
	}

	for {
		if req.State == done {
			break
		}

		// Double the buffer size if it's full
		if len(buf) == readToIndex {
			newBuffer := make([]byte, len(buf)*2)
			copy(newBuffer, buf)
			buf = newBuffer
		}

		// Process any buffered data before doing a read.
		if readToIndex > 0 {
			parsedBytes, err := req.parse(buf[:readToIndex])
			if err != nil {
					return &req, err
			}
			if parsedBytes > 0 {
					// Shift remaining data left.
					copy(buf, buf[parsedBytes:readToIndex])
					readToIndex -= parsedBytes
			} 
			// If the parser now indicates done, break out.
			if req.State == done {
					break
			}
		}

		// Read from the reader into the buffer
		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			break
		} else if err != nil {
			return &req, err
		}
		// Increment the readToIndex to show how many bytes have been read
		readToIndex += n

		// Parse the buffer
		parsedBytes, err := req.parse(buf[:readToIndex])
		if err != nil {
			return &req, err
		}

		// Shift the buffer to the left
		copy(buf, buf[parsedBytes:readToIndex])

		// Decrement the readToIndex to show remaining unparsed bytes
		readToIndex -= parsedBytes
	}

	// Handle any remaining bytes in the buffer
	if readToIndex > 0 && req.State != done {
		_, err := req.parse(buf[:readToIndex])
		if err != nil {
			return &req, err
		}
	}

	return &req, nil
}

// parseRequestLine parses the request line and returns the number of bytes read
func parseRequestLine(b []byte) (*RequestLine, int, error) {
	reqLine := &RequestLine{}
	str := string(b)
	lines := strings.Split(str, "\r\n")
	if len(lines) < 2 {
		return reqLine, 0, nil
	}
	parts := strings.Split(lines[0], " ")
	if len(parts) != 3 {
		return reqLine, 0, fmt.Errorf("invalid number of parts in request line: %s", lines)
	}

	for _, char := range parts[0] {
		if !unicode.IsUpper(char) {
			return reqLine, 0, fmt.Errorf("invalid method: %s", parts[0])
		}
	}

	if parts[2] != "HTTP/1.1" {
		return reqLine, 0, fmt.Errorf("invalid http version: %s", parts[2])
	}

	versionParts := strings.Split(parts[2], "/")

	reqLine.Method = parts[0]
	reqLine.RequestTarget = parts[1]
	reqLine.HttpVersion = versionParts[1]

	return reqLine, len(lines[0]) + 2, nil
}

// parse parses the request line and headers and sets the state to done
func (r *Request) parse(data []byte) (int, error) {
	totalBytesRead := 0
	for r.State != done {
		numBytesPerRead := 0
		switch r.State {
		case initialized:
			reqLine, numBytesPerRead, err := parseRequestLine(data[totalBytesRead:])
			if err != nil {
				return 0, err
			}
	
			if numBytesPerRead == 0 {
				return 0, nil
			}
	
			r.RequestLine = *reqLine
			r.State = parsingHeaders
			totalBytesRead += numBytesPerRead
		case parsingHeaders:
			numBytesPerRead, finished, err := r.Headers.Parse(data[totalBytesRead:])
			if err != nil {
				return 0, err
			}
	
			if finished {
				r.State = done
			}
	
			totalBytesRead += numBytesPerRead
		case done:
			return 0, fmt.Errorf("error: trying to read data in a done state")
		default:
			return 0, fmt.Errorf("error: unknown state")
		}

		if numBytesPerRead == 0 {
			break
		}
	}
	return totalBytesRead, nil
}
