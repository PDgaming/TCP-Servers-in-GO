package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const PORT = ":8080"

func main() {
	fmt.Printf("Starting HTTP Server on port %s\n", PORT)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("error listening: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go handleHttpRequest(conn)
	}
}

func handleHttpRequest(conn net.Conn) {
	defer func() {
		fmt.Printf("Connection closed for %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		if !strings.Contains(err.Error(), "timeout") {
			fmt.Printf("Error reading request line: %v\n", err)
		}
		return
	}

	requestLine = strings.TrimSpace(requestLine)
	parts := strings.Fields(requestLine)

	if len(parts) < 3 {
		fmt.Printf("Received malformed request: %s\n", requestLine)
		return
	}

	method := parts[0]
	path := parts[1]
	protocol := parts[2]

	fmt.Printf("\n--- NEW REQUEST from %s ---\n", conn.RemoteAddr().String())
	fmt.Printf("Method: %s, Path: %s, Protocol: %s\n", method, path, protocol)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.TrimSpace(line) == "" {
			break
		}

		fmt.Printf("Header: %s", line)
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<title>TCP to HTTP</title>
</head>
<body>
	<h1>Success! This is a TCP to HTTP server!</h1>
	<p>You requested the path: <strong>%s</strong></p>
	<p>The connection used <strong>%s</strong> over TCP</p>
</body>
</html>`, path, protocol)

	responseStatus := "HTTP/1.1 200 OK\r\n"

	responseHeaders := fmt.Sprintf("Content-Type: text/html; charset=utf-8\r\n")
	responseHeaders += fmt.Sprintf("Content-Length: %d\r\n", len(htmlBody))
	responseHeaders += "Connection: close\r\n"
	responseHeaders += "\r\n"

	fullResponse := responseStatus + responseHeaders + htmlBody

	_, err = conn.Write([]byte(fullResponse))
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
}
