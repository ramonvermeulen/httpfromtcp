package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ramonvermeulen/httpfromtcp/cmd/server"
	"github.com/ramonvermeulen/httpfromtcp/internal/headers"
	"github.com/ramonvermeulen/httpfromtcp/internal/request"
	"github.com/ramonvermeulen/httpfromtcp/internal/response"
)

const port = 42069

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func main() {
	handler := func(res *response.Writer, req *request.Request) {
		s := response.StatusOK
		h := headers.NewHeaders()
		body := respond200()

		if req.RequestLine.RequestTarget == "/yourproblem" {
			s = response.StatusBadRequest
			body = respond400()
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			s = response.StatusError
			body = respond500()
		} else if after, ok := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin"); ok {
			proxyRes, err := http.Get(fmt.Sprintf("https://httpbin.org%s", after))
			if err != nil {
				s = response.StatusError
				body = respond500()
			} else {
				res.WriteStatusLine(s)
				h.Set("transfer-encoding", proxyRes.Header.Get("transfer-encoding"))
				h.Set("Content-Type", proxyRes.Header.Get("Content-Type"))
				h.Set("Trailer", "X-Content-Length")
				h.Set("Trailer", "X-COntent-SHA256")
				res.WriteHeaders(h)
				fullBody := make([]byte, 0)
				for {
					data := make([]byte, 32)
					n, err := proxyRes.Body.Read(data)
					fullBody = append(fullBody, data[:n]...)
					if err != nil {
						break
					}
					res.WriteChunkedBody(data[:n])
				}
				res.WriteChunkedBodyDone(true)
				hash := sha256.Sum256(fullBody)
				trailer := headers.NewHeaders()
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				trailer.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
				res.WriteTrailers(trailer)
				return
			}
		}

		res.WriteStatusLine(s)
		h.Set("Content-Type", "text/html")
		h.Set("Content-Length", fmt.Sprintf("%d", len(body)))
		res.WriteHeaders(h)
		res.WriteBody(body)
	}
	server, err := server.Serve(handler, port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
