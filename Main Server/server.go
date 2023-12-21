package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Maximum limit of accepting client request at a time.
const maxConnection int = 10

func main() {
	//Checks if the user have given the proper argument or not.
	if len(os.Args) != 2 {
		fmt.Println("Wrong argument. It should be like: go run servername <port>>")
		return
	}

	port_number := os.Args[1]

	/*This function attempts to create a network listener for a TCP network service.
	If the operation is successful, it returns a listener object and err is set to nil,
	which indicats that no error occurred.
	If an error occurs: such as if the specified port is already in use or there's a network configuration problem,
	err is assigned an error value that describes the issue.*/
	listener, err := net.Listen("tcp", ":"+port_number)
	if err != nil {
		fmt.Println("Error listening connection: ", err)
		os.Exit(1)
	}
	//To close the listener when the server is terminated
	defer listener.Close()

	fmt.Printf("Server is listening on :%s\n", port_number)

	/*A WaitGroup is basically a struct type defined in the sync package.
	- The variable wg is a new wait group instance that will be used to synchronize
	and wait for the completion of groups of go routines.*/
	var wg sync.WaitGroup

	//Buffer channel to manage the connection to its predefined limit.
	semaphore := make(chan struct{}, maxConnection)

	//To listen to the continuous connections which will be coming to the server. Max handle limit = 10
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		// Before trying to send to the semaphore, check if it's at capacity.
		if len(semaphore) == maxConnection {
			fmt.Println("Connection is waiting...") // This will print when the semaphore is full, and the connection has to wait.
		}
		semaphore <- struct{}{} // Block if maxConnections is reached
		wg.Add(1)
		go func() {
			handleRequest(conn)
			<-semaphore // Release a spot
			wg.Done()
		}()
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close() //close connection on exit

	// This uses net/http to parse the request which is more reliable.
	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Error reading request: %v", err)
		}
		sendErrorResponse(conn, http.StatusBadRequest, "Bad Request")
		return
	}

	switch req.Method {
	case "GET":
		handleGetRequest(conn, req)
	case "POST":
		handlePostRequest(conn, req)
	default:
		sendErrorResponse(conn, http.StatusNotImplemented, "Not Implemented")
	}
}

func handleGetRequest(conn net.Conn, req *http.Request) {
	// Extract the resource path from the request URL
	resourcePath := req.URL.Path[1:] // remove the leading '/'

	// Determine the content type of the requested file
	contentType := getContentType(resourcePath)

	// Prevent directory traversal attacks
	if strings.Contains(resourcePath, "..") {
		sendContentErrorResponse(conn, http.StatusBadRequest, contentType, "Bad Request")
		return
	}

	// For binary files or when the content type cannot be determined, send a Bad Request error
	if contentType == "application/octet-stream" {
		sendContentErrorResponse(conn, http.StatusBadRequest, contentType, "Bad Request")
		return
	}

	// Open the requested file
	requestedFile, err := os.Open(resourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			sendContentErrorResponse(conn, http.StatusNotFound, contentType, "Not Found")
		} else {
			sendErrorResponse(conn, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	defer requestedFile.Close()

	// Send the response header and the file content
	sendGETResponse(conn, http.StatusOK, contentType, requestedFile)
}

func handlePostRequest(conn net.Conn, req *http.Request) {
	// Create the file
	resourcePath := req.URL.Path[1:] // remove the leading '/'

	// Determine the content type of the requested file
	contentType := getContentType(resourcePath)
	// Ensure the path is safe
	if strings.Contains(req.URL.Path, "..") {
		sendContentErrorResponse(conn, http.StatusBadRequest, contentType, "Bad Request")
		return
	}

	// For binary files or when the content type cannot be determined, send a Bad Request error
	if contentType == "application/octet-stream" {
		sendContentErrorResponse(conn, http.StatusBadRequest, contentType, "Bad Request")
		return
	}

	file, err := os.Create(resourcePath)
	if err != nil {
		sendErrorResponse(conn, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	defer file.Close()

	// Write the body to file
	_, err = io.Copy(file, req.Body)
	if err != nil {
		sendErrorResponse(conn, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Send a success response
	sendPostResponse(conn, http.StatusOK, contentType, "File uploaded successfully.")
}

// For processing specific file extension to client.
func getContentType(resource_path string) string {
	extension := strings.ToLower(filepath.Ext(resource_path))

	switch extension {
	case ".html":
		return "text/html"
	case ".txt":
		return "text/plain"
	case ".gif":
		return "image/gif"
	case ".jpeg", ".jpg":
		return "image/jpeg"
	case ".css":
		return "text/css"
	default:
		return "application/octet-stream"
	}
	/*If the file extension doesn't match any of the predefined cases,
	the default case is triggered, and
	it returns "application/octet-stream."
	This is a generic content type used for unknown or binary file types,
	indicating that the file should be treated as a binary stream.*/
}

/*
For sending an HTTP response with an error status code and
a status notification message to the client.
Sprintf formats according to a format specifier and returns the resulting string.
\r\n\r\n to indicate the end of the headers and the start of the response content.
The second occurrence of statusNotification in the reponse section is used to provide a more detailed explanation in the response body,
which is separated from the status line by a blank line.
*/
func sendErrorResponse(conn net.Conn, statusCode int, statusNotification string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n%s", statusCode, statusNotification, statusNotification)
	//sends the response header to the client.
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending error response: %v\n", err)
		conn.Close()
	}
}

func sendContentErrorResponse(conn net.Conn, statusCode int, contentType string, statusNotification string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\nContent_Type: %s\r\n\r\n%s", statusCode, statusNotification, contentType, statusNotification)
	//sends the response header to the client.
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending error response: %v\n", err)
		conn.Close()
	}
}

/*
For sending an HTTP response header, including the status code, content type,
and response headers, followed by the actual content to the client.
An io.Reader interface representing the content (file or data) to be sent in the response.
*/
func sendGETResponse(conn net.Conn, statusCode int, content_type string, requested_file *os.File) {
	// Read the file content into a byte slice
	/*fileContent, err := io.ReadAll(requested_file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		sendErrorResponse(conn, http.StatusInternalServerError, "Internal Server Error")
		return
	}*/
	// Get file info to determine the content length
	fileInfo, err := requested_file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info: %v\n", err)
		sendErrorResponse(conn, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	//constructs the response header.
	response := fmt.Sprintf("HTTP/1.1 %d OK\r\n", statusCode)
	response += fmt.Sprintf("Content-Type: %s\r\n", content_type)
	response += fmt.Sprintf("Content-Length: %d\r\n", fileInfo.Size())
	response += "Server: Go\r\n"
	response += "Connection: close\r\n"
	response += "\r\n"
	//sends the response header to the client.
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending GET response: %v\n", err)
		conn.Close()
	}

	io.Copy(conn, requested_file) //copies the content from requested_file to the network connection conn.
	//This is used to send the actual content to the client.
}

// New function to send a simple response after a POST request
func sendPostResponse(conn net.Conn, statusCode int, contentType string, message string) {
	//constructs the response header.
	response := fmt.Sprintf("HTTP/1.1 %d OK\r\n", statusCode)
	response += fmt.Sprintf("Content-Type: %s\r\n", contentType)
	response += fmt.Sprintf("Content-Length: %d\r\n", len(message))
	response += "Server: Go\r\n"
	response += "Connection: close\r\n"
	response += "\r\n"

	// Write the headers to the client
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending POST response: %v\n", err)
		conn.Close()
	}

	// Write the headers to the client
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Printf("Error sending POST response: %v\n", err)
		conn.Close()
	}
}
