package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string, 1)
	content := make([]byte, 8)

	go func() {
		defer close(c)
		defer f.Close()
		var line string
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
	msg, err := os.Open("messages.txt")
	if err != nil {
		panic(fmt.Errorf("error occurred while opening file: %w", err))
	}

	c := getLinesChannel(msg)
	for line := range c {
		fmt.Printf("read: %s\n", line)
	}
}
