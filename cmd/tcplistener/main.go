package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string, 1)
	content := make([]byte, 8)
	var line string

	go func() {
		defer func() {
			if line != "" {
				c <- line
			}
			close(c)
			fmt.Println("closed connection")
			f.Close()
		}()
		for {
			n, err := f.Read(content)
			content := content[:n]
			if err != nil {
				break
			}

			parts := strings.Split(string(content), "\n")
			if len(parts) > 1 {
				for i := range len(parts) {
					if i == len(parts)-1 {
						line += parts[i]
						break
					}
					line += parts[i]
					c <- line
					line = ""
				}
			} else {
				line += parts[0]
			}
		}
	}()

	return c
}

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

		c := getLinesChannel(conn)
		for line := range c {
			fmt.Printf("read: %s\n", line)
		}
	}
}
