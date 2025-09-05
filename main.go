package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	msg, err := os.Open("messages.txt")
	if err != nil {
		panic(fmt.Errorf("error occurred while opening file: %w", err))
	}

	content := make([]byte, 8)
	n, err := msg.Read(content)
	var line string

	for err == nil {
		parts := strings.Split(string(content[:n]), "\n")
		if len(parts) > 1 {
			for i := range len(parts) {
				if i == len(parts)-1 {
					line += parts[i]
					break
				}
				line += parts[i]
				fmt.Printf("read: %s\n", line)
				line = ""
			}
		} else {
			line += parts[0]
		}
		n, err = msg.Read(content)
	}

	if err == io.EOF {
		os.Exit(0)
	} else {
		panic(fmt.Errorf("error occurred while reading file: %w", err))
	}
}
