package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/isotronic/httpfromtcp/internal/headers"
)

type Writer struct {
	writer io.Writer
	state  writerState
}

type writerState int

const (
	writingStatusLine writerState = iota
	writingHeaders
	writingBody
)

type StatusCode int

const (
	StatusOK									StatusCode = 200
	StatusBadRequest					StatusCode = 400
	StatusInternalServerError	StatusCode = 500
)

// NewWriter creates a new Writer with the given io.Writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{
			writer: w,
			state: writingStatusLine,
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Add("Content-Length", strconv.Itoa(contentLen))
	h.Add("Connection", "close")
	h.Add("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writingStatusLine {
		return fmt.Errorf("error: cannot write status line")
	}
	defer func() {
		w.state = writingHeaders
	}()
	switch statusCode {
	case StatusOK:
		_, err := w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
		return nil
	case StatusBadRequest:
		_, err := w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
		return nil
	case StatusInternalServerError:
		_, err := w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
		return nil
	default:
		return nil
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writingHeaders {
		return fmt.Errorf("error: cannot write headers before writing status line")
	}
	defer func() {
		w.state = writingBody
	}()
	responseHeaders := ""
	for header := range headers {
		responseHeaders += header + ": " + headers[header] + "\r\n"
	}
	_, err := w.writer.Write([]byte(responseHeaders + "\r\n"))
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("error: cannot write body before writing headers")
	}

	return w.writer.Write(p)
}