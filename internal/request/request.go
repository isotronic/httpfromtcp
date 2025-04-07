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
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	reqLine, err := parseRequestLine(data)
	if err != nil {
		return nil, err
	}

	req := Request{
		RequestLine: *reqLine,
	}

	return &req, nil
}

func parseRequestLine(b []byte) (*RequestLine, error) {
	reqLine := &RequestLine{}
	str := string(b)
	line := strings.Split(str, "\r\n")[0]
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return reqLine , fmt.Errorf("invalid number of parts in request line: %s", line)
	}

	for _, char := range parts[0] {
		if !unicode.IsUpper(char) {
			return reqLine, fmt.Errorf("invalid method: %s", parts[0])
		}
	}

	if parts[2] != "HTTP/1.1" {
		return reqLine, fmt.Errorf("invalid http version: %s", parts[2])
	}

	versionParts := strings.Split(parts[2], "/")

	reqLine.Method = parts[0]
	reqLine.RequestTarget = parts[1]
	reqLine.HttpVersion = versionParts[1]

	return reqLine, nil
}