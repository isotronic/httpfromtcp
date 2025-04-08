package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	initialized = iota
	done
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := Request{
		State: initialized,
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

// parse parses the request line and sets the state to done
func (r *Request) parse(data []byte) (int, error) {
	if r.State == initialized {
		reqLine, numBytesPerRead, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if numBytesPerRead == 0 {
			return 0, nil
		}

		r.RequestLine = *reqLine
		r.State = done
		return numBytesPerRead, nil
	} else if r.State == done {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	} else {
		return 0, fmt.Errorf("error: unknown state")
	}
}
