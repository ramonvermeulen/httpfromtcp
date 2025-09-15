package response

import (
	"fmt"
	"io"

	"github.com/ramonvermeulen/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK         StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusError      StatusCode = 500
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case StatusBadRequest:
		w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusError:
		w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		fmt.Fprintf(w.writer, "HTTP/1.1 %d\r\n", statusCode)
	}

	return nil
}

func (w *Writer) WriteHeaders(hdrs headers.Headers) error {
	for key, value := range hdrs {
		fmt.Fprintf(w.writer, "%s: %s\r\n", key, value)
	}
	fmt.Fprintf(w.writer, "\r\n")
	return nil
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	n, err := w.writer.Write(body)
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	fmt.Fprintf(w.writer, "%x\r\n", len(p))
	fmt.Fprintf(w.writer, "%s\r\n", p)
	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	endChunk := []byte("0\r\n\r\n")
	w.writer.Write(endChunk)
	return len(endChunk), nil
}
