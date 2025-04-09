package response

import (
	"io"
	"strconv"

	"github.com/isotronic/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK									StatusCode = 200
	StatusBadRequest					StatusCode = 400
	StatusInternalServerError	StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
		return nil
	case StatusBadRequest:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
		return nil
	case StatusInternalServerError:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
		return nil
	default:
		return nil
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.Headers{
		"Content-Length": strconv.Itoa(contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	responseHeaders := ""
	for header := range headers {
		responseHeaders += header + ": " + headers[header] + "\r\n"
	}
	_, err := w.Write([]byte(responseHeaders + "\r\n"))
	if err != nil {
		return err
	}

	return nil
}