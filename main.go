package main

import (
	"fmt"
	"net"
	"net/http"
	"bufio"
	"os/exec"
)

var (
	urlQueue = make(chan string, 100) // Channel as a queue with buffer
)

func submitHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}
	// Send the URL to the queue
	urlQueue <- url
	fmt.Fprintf(w, "URL queued: %s\n", url)
}

func processUrls() {
	for url := range urlQueue {
		fmt.Printf("Processing URL: %s\n", url)
		// Source .bashrc and call your bash function
		cmd := exec.Command("/bin/bash", "-c", "source ~/.bashrc >/dev/null 2>&1; fake-foost '"+url+"'")

		// Get the pipe for the standard output of the command
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Printf("Error obtaining stdout: %s\n", err)
			continue
		}

		// Similarly, get the pipe for the standard error of the command
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			fmt.Printf("Error obtaining stderr: %s\n", err)
			continue
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error starting command: %s\n", err)
			continue
		}

		// Use a scanner to read the command's standard output line by line
		stdoutScanner := bufio.NewScanner(stdoutPipe)
		go func() {
			for stdoutScanner.Scan() {
				fmt.Printf("STDOUT: %s\n", stdoutScanner.Text())
			}
		}()

		// Do the same for the standard error
		stderrScanner := bufio.NewScanner(stderrPipe)
		go func() {
			for stderrScanner.Scan() {
				fmt.Printf("STDERR: %s\n", stderrScanner.Text())
			}
		}()

		// Wait for the command to finish
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("Command finished with error: %s\n", err)
		}
	}
}


func main() {
	listener, err := net.Listen("tcp", "localhost:0") // 0 means an available port will be chosen
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		return
	}
	defer listener.Close()

	go processUrls()

	http.HandleFunc("/submit", submitHandler)
	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Server starting on port %d...\n", port)
	fmt.Printf("Submit URLs with: curl http://localhost:%d/submit?url=<URL>\n", port)

	if err := http.Serve(listener, nil); err != nil {
		fmt.Printf("Error running server: %s\n", err)
	}
}
