package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		targetURL = "http://localhost:8080"
	}

	email := "sarah.chen@aidynamics.com"
	text := "What is the pricing for TormentNexus?"

	fmt.Printf("Starting UAT verification for: %s\n", targetURL)
	fmt.Printf("Simulating inbound from %s: %s\n", email, text)

	data := url.Values{}
	data.Set("email", email)
	data.Set("text", text)

	resp, err := http.PostForm(fmt.Sprintf("%s/api/v1/test/simulate_inbound", targetURL), data)
	if err != nil {
		fmt.Printf("UAT verification failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("UAT verification failed: status %d, body: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("UAT verification successful.\nResponse: %s\n", string(body))

	if !strings.Contains(string(body), "Autonomous reply") {
		fmt.Println("UAT verification failed: Missing autonomous reply in response.")
		os.Exit(1)
	}
}
