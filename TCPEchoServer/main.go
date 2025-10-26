package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const (
	PORT        = ":8080"
	BUFFER_SIZE = 1024
)

func main() {
	fmt.Printf("Starting TCP Echo Server on port %s\n", PORT)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("Error Listening: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("New client connected from %s\n", conn.RemoteAddr().String())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Printf("Connection closed for %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()

	buffer := make([]byte, BUFFER_SIZE)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			fmt.Printf("Error reading from connection: %v\n", err)
			return
		}

		if n == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		message := string(buffer[:n])
		fmt.Printf("Received [%d bytes]: %s", n, message)

		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Error writing to connection: %v\n", err)
			return
		}

		fmt.Printf("Echoed [%d bytes] back.\n", n)
	}
}
