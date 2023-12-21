package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	//("Enter the command, i.e: localhost:8080 /file.txt receive 20")
	serverURL := os.Args[1]     //"localhost:8080" // Replace with server's URL
	service := os.Args[2]       //client's request for a particular file
	process := os.Args[3]       //to send the data or receive from the server.
	client_number := os.Args[4] //enter the number of clients that will send concurrent requests to the server.

	c_n, _ := strconv.Atoi(client_number)
	var i int

	for i = 0; i < c_n; i++ {
		if process == "receive" {
			// Construct the URL for the specific file extension
			fileURL := serverURL + service
			// Send a GET request for files with the specified extension
			getResponse, err := http.Get(fileURL)
			if err != nil {
				fmt.Printf("GET request for %s failed: %v\n Status: Bad Request (400)\n", service, err)
				return
			}
			defer getResponse.Body.Close()

			// Read and display the response from the server
			fileContent, err := ioutil.ReadAll(getResponse.Body)
			if err != nil {
				fmt.Printf("Error reading %s response: %v\n", service, err)
			}

			if getResponse.StatusCode != http.StatusOK {
				fmt.Printf("GET request for %s returned status: %s (%d)\n", service, string(fileContent), getResponse.StatusCode)
			} else {
				fmt.Printf("GET request for %s returned status: OK (%d)\n", service, getResponse.StatusCode)
			}

			fmt.Printf("File Content: %s\n", service)
			fmt.Println(string(fileContent))
		} else if process == "send" {
			fileURL := serverURL + service
			if strings.HasPrefix(service, "/") {
				file_send := service[1:]

				// Openning the file to send
				file, err := os.Open(file_send)
				if err != nil {
					fmt.Printf("Error opening file: %v\n", err)
					return
				}
				defer file.Close()

				// Read the contents of the file
				fileContents, err := io.ReadAll(file)
				if err != nil {
					fmt.Printf("Error reading file: %v\n", err)
					return
				}
				fmt.Println("Client00.....")
				// Send a POST request with the file contents
				getResponse, err := http.Post(fileURL, "text/plain", bytes.NewBuffer(fileContents)) // or bytes.NewReader
				fmt.Println("Client.....")
				if err != nil {
					fmt.Printf("POST request failed: %v\n", err)
					return
				}
				defer getResponse.Body.Close()

				// Read and display the response from the server
				fileContent, err := io.ReadAll(getResponse.Body)
				if err != nil {
					fmt.Printf("Error reading %s response: %v\n", service, err)
				}

				if getResponse.StatusCode != http.StatusOK {
					fmt.Printf("POST request for %s returned status: %s (%d)\n", service, string(fileContent), getResponse.StatusCode)
				} else {
					fmt.Printf("POST request for %s returned status: OK (%d)\n", service, getResponse.StatusCode)
				}

				fmt.Printf("File Content: %s\n", service)
				fmt.Println(string(fileContent))
			}
		}
	}
}
