// counters.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	counterMu   sync.Mutex
	counterFile = "./chatterCount.json"
)

// incrementChatterCounter reads the current count, increments, and writes it back.
func incrementChatterCounter() error {
	counterMu.Lock()
	defer counterMu.Unlock()

	// ensure file exists
	if _, err := os.Stat(counterFile); os.IsNotExist(err) {
		if err := os.WriteFile(counterFile, []byte(`{"count":0}`), 0644); err != nil {
			return fmt.Errorf("create counter file: %w", err)
		}
	}

	// read
	data, err := os.ReadFile(counterFile)
	if err != nil {
		return fmt.Errorf("read counter: %w", err)
	}
	var obj struct{ Count int `json:"count"` }
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("parse counter JSON: %w", err)
	}

	// increment
	obj.Count++

	// write back
	out, _ := json.MarshalIndent(obj, "", "  ")
	if err := os.WriteFile(counterFile, out, 0644); err != nil {
		return fmt.Errorf("write counter: %w", err)
	}
	return nil
}

// getChatterCount returns the current count.
func getChatterCount() (int, error) {
	counterMu.Lock()
	defer counterMu.Unlock()

	data, err := os.ReadFile(counterFile)
	if err != nil {
		return 0, err
	}
	var obj struct{ Count int `json:"count"` }
	if err := json.Unmarshal(data, &obj); err != nil {
		return 0, err
	}
	return obj.Count, nil
}
