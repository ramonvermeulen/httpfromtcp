package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	msg, err := os.Open("messages.txt")
	if err != nil {
		panic(fmt.Errorf("error occurred while opening file: %w", err))
	}

	content := make([]byte, 8)
	n, err := msg.Read(content)
	for err == nil {
		fmt.Printf("read: %s\n", content[:n])
		n, err = msg.Read(content)
	}
	if err == io.EOF {
		os.Exit(0)
	} else {
		panic(fmt.Errorf("error occurred while reading file: %w", err))
	}
}
