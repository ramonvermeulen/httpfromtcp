package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ramonvermeulen/httpfromtcp/internal/request"
)

func main() {
	lis, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("accepted connection")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error processing request")
		}
		fmt.Printf("Request line\n")
		fmt.Printf("  - Method: %s\n", request.RequestLine.Method)
		fmt.Printf("  - Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("  - Version: %s\n", request.RequestLine.HTTPVersion)
		fmt.Printf("Headers:\n")
		for key, value := range request.Headers {
			fmt.Printf("  - %s: %s\n", key, value)
		}
		fmt.Printf("Body:\n%s", request.Body)
	}
}
