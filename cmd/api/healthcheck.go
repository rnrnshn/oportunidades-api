package main

import (
	"fmt"
	"net/http"
	"time"
)

// The runtime image has no shell or curl, so container healthchecks run the binary itself.
func runHealthcheck(port string) error {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %d", resp.StatusCode)
	}
	return nil
}
