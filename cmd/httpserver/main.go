package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ramonvermeulen/httpfromtcp/cmd/server"
	"github.com/ramonvermeulen/httpfromtcp/internal/request"
	"github.com/ramonvermeulen/httpfromtcp/internal/response"
)

const port = 42069

func main() {
	handler := func(res io.Writer, req *request.Request) *server.HandlerError {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your problem is not my problem.\n",
			}
		}
		if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				StatusCode: response.StatusError,
				Message:    "Woopsie, my bad\n",
			}
		}

		res.Write([]byte("Hello, world!\n"))
		return nil
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
