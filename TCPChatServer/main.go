package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const PORT = ":8080"
const BUFFER_SIZE = 1024

var clients = make(map[string]net.Conn)

var mutex = &sync.Mutex{}

func main() {
	fmt.Printf("--- Starting TCP Chat Server on port %s ---\n", PORT)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("[FATAL ERROR]: Could not listen on port %s: %v\n", PORT, err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server initialized. Waiting for connections...")

	for {
		fmt.Println("[DEBUG]: Server is blocking on listener.Accept()...")

		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[ERROR]: Error accepting connection: %v\n", err)
			continue
		}

		fmt.Printf("[DEBUG]: Accepted connection from %s. Starting handler goroutine.\n", conn.RemoteAddr().String())

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	addClient(conn)

	defer func() {
		removeClient(conn)
		conn.Close()
	}()

	buffer := make([]byte, BUFFER_SIZE)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			fmt.Printf("[ERROR]: Error reading from %s: %v\n", addr, err)
			return
		}

		if n > 0 {
			message := string(buffer[:n])
			trimmedMessage := strings.TrimSpace(message)

			if trimmedMessage != "" {
				broadcastMessage(addr, addr, trimmedMessage)
			}
		}
	}
}

func addClient(conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	addr := conn.RemoteAddr().String()
	clients[addr] = conn

	welcomeMessage := fmt.Sprintf("Welcome to the chat, %s!\n", addr)

	_, err := conn.Write([]byte(welcomeMessage))
	if err != nil {
		fmt.Printf("[ERROR]: Failed to write welcome message to %s: %v\n", addr, err)
	}

	joinMessage := fmt.Sprintf("A new client (%s) has joined the chat.>", addr)
	broadcastToOthers(addr, "SERVER", joinMessage)

	fmt.Printf("Active Connections: %d\n", len(clients))
}

func removeClient(conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	addr := conn.RemoteAddr().String()
	delete(clients, addr)

	leaveMessage := fmt.Sprintf("Client (%s) has left the chat.>", addr)
	broadcastToOthers(addr, "SERVER", leaveMessage)

	fmt.Printf("Active Connections: %d\n", len(clients))
}

func broadcastToOthers(senderAddr string, senderName string, message string) {
	timestamp := time.Now().Format("15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, senderName, message)

	fmt.Printf("Broadcasting: %s", formattedMessage)

	for addr, conn := range clients {
		if addr == senderAddr {
			_, err := conn.Write([]byte(">"))
			if err != nil {
				fmt.Printf("[ERROR]: Error %v\n", err)
			}
			continue
		}

		_, err := conn.Write([]byte(fmt.Sprintf("%s>", formattedMessage)))
		if err != nil {
			fmt.Printf("[ERROR]: Error sending to client %s: %v. Client likely disconnected.\n", addr, err)
		}
	}
}

func broadcastMessage(senderAddr string, senderName string, message string) {
	mutex.Lock()
	defer mutex.Unlock()

	timestamp := time.Now().Format("15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, senderName, message)

	fmt.Printf("Broadcasting: %s", formattedMessage)

	for addr, conn := range clients {
		if addr == senderAddr {
			_, err := conn.Write([]byte(">"))
			if err != nil {
				fmt.Printf("[ERROR]: Error %v\n", err)
			}
			continue
		}

		_, err := conn.Write([]byte(fmt.Sprintf("%s>", formattedMessage)))
		if err != nil {
			fmt.Printf("[ERROR]: Error sending to client %s: %v. Client likely disconnected.\n", addr, err)
		}
	}
}
