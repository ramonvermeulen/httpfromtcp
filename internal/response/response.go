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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case StatusBadRequest:
		w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusError:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		fmt.Fprintf(w, "HTTP/1.1 %d\r\n", statusCode)
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, hdrs headers.Headers) error {
	for key, value := range hdrs {
		fmt.Fprintf(w, "%s: %s\r\n", key, value)
	}
	return nil
}
