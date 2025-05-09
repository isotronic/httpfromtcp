package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/isotronic/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine 		RequestLine
	Headers     		headers.Headers
	Body        		[]byte

	bodyReadLength 	int
	state       		RequestState
}

type RequestLine struct {
	HttpVersion   	string
	RequestTarget 	string
	Method        	string
}

type RequestState int

const (
	initialized RequestState = iota
	parsingHeaders
	parsingBody
	done
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := Request{
		state:   initialized,
		Headers: headers.Headers{},
	}

	for {
		if req.state == done {
			break
		}

		// Double the buffer size if it's full
		if len(buf) == readToIndex {
			newBuffer := make([]byte, len(buf)*2)
			copy(newBuffer, buf)
			buf = newBuffer
		}

		// Always attempt to parse what we already have
		parsedBytes, err := req.parse(buf[:readToIndex])
		if err != nil {
			return &req, err
		}
		if parsedBytes > 0 {
			// Shift remaining data left.
			copy(buf, buf[parsedBytes:readToIndex])
			readToIndex -= parsedBytes

			// If parsing completed the request then break.
			if req.state == done {
				break
			}
		} else {
			// No progress was made by parsing – try to read more data.
			n, err := reader.Read(buf[readToIndex:])
			if err == io.EOF {
				break
			} else if err != nil {
				return &req, err
			}
			readToIndex += n
		}
	}

	// Handle any remaining bytes in the buffer
	if readToIndex > 0 && req.state != done {
		_, err := req.parse(buf[:readToIndex])
		if err != nil {
			return &req, err
		}
	}

	length, ok := req.Headers["content-length"]
	if ok {
		contentLength, err := strconv.Atoi(length)
		if err != nil {
			return &req, err
		}
		if contentLength != len(req.Body) {
			return &req, fmt.Errorf("error: content length does not match body length")
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
		return reqLine, 0, fmt.Errorf("invalid number of parts in request line: %s", lines[0])
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
	for r.state != done {
		numBytesPerRead := 0
		switch r.state {
		case initialized:
			reqLine, numBytesPerRead, err := parseRequestLine(data[totalBytesRead:])
			if err != nil {
				return 0, err
			}
	
			if numBytesPerRead == 0 {
				return 0, nil
			}
	
			r.RequestLine = *reqLine
			r.state = parsingHeaders
			totalBytesRead += numBytesPerRead
		case parsingHeaders:
			numBytesPerRead, finished, err := r.Headers.Parse(data[totalBytesRead:])
			if err != nil {
				return 0, err
			}
	
			if finished {
        // If no Content-Length header is present, then there is no body.
        if _, ok := r.Headers["content-length"]; !ok {
          r.state = done
        } else {
          r.state = parsingBody
        }
			}
	
			totalBytesRead += numBytesPerRead
		case parsingBody:
			length, ok := r.Headers["content-length"]
			if !ok {
				r.state = done
				return len(data), nil
			}
			contentLength, err := strconv.Atoi(length)
			if err != nil {
				return 0, err
			}
			if contentLength == 0 {
				r.state = done
				return len(data), nil
			}
			r.Body = append(r.Body, data...)
			r.bodyReadLength += len(data)

			if contentLength < r.bodyReadLength {
				return 0, fmt.Errorf("error: content length is less than body length")
			}
			if contentLength == r.bodyReadLength {
				r.state = done
			}

			return len(data), nil
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
