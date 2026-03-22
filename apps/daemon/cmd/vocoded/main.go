package main

import (
	"bufio"
	"io"
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stderr)
	log.Println("vocoded starting...")

	reader := bufio.NewReader(os.Stdin)
	for {
		_, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				log.Println("stdin closed, shutting down")
				return
			}
			log.Printf("stdin read error: %v", err)
			return
		}
	}
}
