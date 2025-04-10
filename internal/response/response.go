package response

import (
	"fmt"
	"io"
	"net/http"
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
	writingTrailers
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

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writingTrailers {
		return fmt.Errorf("error: cannot write trailers before writing body")
	}
	responseTrailers := ""
	for header := range h {
		responseTrailers += header + ": " + h[header] + "\r\n"
	}
	_, err := w.writer.Write([]byte(responseTrailers + "\r\n"))
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("error: cannot write body before writing headers")
	}
	defer func() {
		w.state = writingTrailers
	}()

	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("error: cannot write chunked body before writing headers")
	}
	hex := fmt.Sprintf("%x", len(p))
	str := hex + "\r\n" + string(p) + "\r\n"
	n, err := w.writer.Write([]byte(str))

	if f, ok := w.writer.(http.Flusher); ok {
		f.Flush()
	}

	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	defer func() {
		w.state = writingTrailers
	}()

	n, err := w.writer.Write([]byte("0\r\n"))

	if f, ok := w.writer.(http.Flusher); ok {
		f.Flush()
	}

	return n, err
}