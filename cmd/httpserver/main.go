package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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
		}
		if req.RequestLine.RequestTarget == "/myproblem" {
			s = response.StatusError
			body = respond500()
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
