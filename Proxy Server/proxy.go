package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: proxy <port>")
		return
	}

	port := os.Args[1]
	serverURL := os.Args[2] //"localhost:8080" // Replace with server's URL

	//Listen for client connection
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Error listening on port %s: %v\n", port, err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Proxy server started on port %s\n", port)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go handleClientRequest(clientConn, serverURL)
	}
}

func handleClientRequest(clientConn net.Conn, serverURL string) {
	defer clientConn.Close()

	// Read the incoming request from the client
	req, err := http.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Error reading request from client: %v", err)
		}
		sendErrorResponse(clientConn, http.StatusBadRequest, "Bad Request")
		return
	}

	// Check if the request method is GET
	if req.Method != http.MethodGet {
		sendErrorResponse(clientConn, http.StatusNotImplemented, "Not Implemented")
		return
	} else {
		handleGetRequest(clientConn, req, serverURL)
	}
}

func handleGetRequest(clientConn net.Conn, req *http.Request, serverURL string) {
	serverConn, err := net.Dial("tcp", serverURL)
	if err != nil {
		fmt.Printf("Error connecting to target server at %s: %v\n", serverURL, err)
		return
	}
	defer serverConn.Close()

	// Forward the request to the server by writing it to the server connection
	err = req.Write(serverConn)
	if err != nil {
		fmt.Printf("Error writing request to server: %v\n", err)
		return
	}

	// Read the response from the server
	serverResponse, err := http.ReadResponse(bufio.NewReader(serverConn), req)
	if err != nil {
		fmt.Printf("Error reading response from server: %v\n", err)
		return
	}
	defer serverResponse.Body.Close()

	// Copy the server's response headers and body back to the original client
	serverResponse.Write(clientConn)
}

func sendErrorResponse(conn net.Conn, statusCode int, statusNotification string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n%s", statusCode, statusNotification, statusNotification)
	//sends the server response to the client.
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending server response: %v\n", err)
		conn.Close()
	}
}
