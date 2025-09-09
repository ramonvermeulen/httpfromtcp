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

		rl, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error processing request")
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %+v\n", rl.RequestLine.Method)
		fmt.Printf("- Target: %+v\n", rl.RequestLine.RequestTarget)
		fmt.Printf("- Version: %+v\n", rl.RequestLine.HTTPVersion)
	}
}
