package main_test

import (
	"bufio"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server Output", func() {
	var serverCmd *exec.Cmd
	var serverOutput []string
	var serverErrOutput []string
	var outputLock sync.Mutex
	var wg sync.WaitGroup

	BeforeEach(func() {
		serverCmd = exec.Command("go", "run", "main.go")

		stdoutPipe, err := serverCmd.StdoutPipe()
		Expect(err).NotTo(HaveOccurred())

		stderrPipe, err := serverCmd.StderrPipe()
		Expect(err).NotTo(HaveOccurred())

		err = serverCmd.Start()
		Expect(err).NotTo(HaveOccurred())

		wg.Add(2)

		// Capture standard output
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				text := scanner.Text()
				outputLock.Lock()
				serverOutput = append(serverOutput, text)
				fmt.Println(text)
				outputLock.Unlock()
			}
		}()

		// Capture standard error output
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				text := scanner.Text()
				outputLock.Lock()
				serverErrOutput = append(serverErrOutput, text)
				outputLock.Unlock()
			}
		}()

		// Wait a bit for the server to initialize
		// (Adjust the duration according to your server's startup time)
		Eventually(func() bool {
			outputLock.Lock()
			defer outputLock.Unlock()
			return len(serverOutput) > 0 // Assuming the server prints something upon starting
		}, "20s", "100ms").Should(BeTrue(), "Server did not start in time")
	})

	AfterEach(func() {
		if serverCmd != nil && serverCmd.Process != nil {
			serverCmd.Process.Kill()
			serverCmd.Wait()
		}
		wg.Wait()
	})

	It("should respond with 'Processing URL' and 'Processed URL' messages", func() {
		// Define the fake URL to be submitted
		fakeURL := "http://example.com/fakevideo"

		// Dynamically retrieve the port on which the server started.
		// Assuming your server prints "Server starting on port XXXXX..."
		// and you have captured this output to extract the port:
		portRegex := regexp.MustCompile("Server starting on port (\\d+)")
		var port string
		for _, line := range serverOutput {
			if matches := portRegex.FindStringSubmatch(line); matches != nil {
				port = matches[1]
				break
			}
		}
		Expect(port).NotTo(BeEmpty(), "Failed to find server's port in its output")
		// Construct the URL for submitting the fake URL to the server
		submitURL := fmt.Sprintf("http://localhost:%s/submit?url=%s", port, fakeURL)

		// Simulate submitting the URL to the server
		_, err := http.Get(submitURL)
		Expect(err).NotTo(HaveOccurred(), "Failed to submit URL to server")

		// Use Eventually to wait for the server's response to include the processing messages
		var outputCombined string
		Eventually(func() bool {
			outputLock.Lock()
			defer outputLock.Unlock()
			outputCombined = strings.Join(serverOutput, "\n") + "\n" + strings.Join(serverErrOutput, "\n")

			processingMsg := fmt.Sprintf("Processing URL: %s", fakeURL)
			processedMsg := fmt.Sprintf("Processed: %s", fakeURL)
			fmt.Println(processedMsg)
			// Check if both processing and processed messages are in the output
			return strings.Contains(outputCombined, processingMsg) && strings.Contains(outputCombined, processedMsg)
		}, "20s", "1s").Should(BeTrue(), "Server did not process the URL as expected within the timeout period")
	})

})
