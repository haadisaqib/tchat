// client_id.go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const IDFile = ".tchat_id"

// LoadOrCreateClientID checks for an existing ID file or creates one if it doesn't exist
func LoadOrCreateClientID() (int, error) {
	path, err := filepath.Abs(IDFile)
	if err != nil {
		return 0, fmt.Errorf("could not resolve ID file path: %w", err)
	}

	// If file exists, read it
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return 0, fmt.Errorf("failed to read ID file: %w", err)
		}
		idStr := string(data)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return 0, fmt.Errorf("invalid ID format: %w", err)
		}
		return id, nil
	}

	// File doesn't exist, generate and save one
	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(999999-100000) + 100000
	idStr := strconv.Itoa(id)
	err = os.WriteFile(path, []byte(idStr), 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to write ID file: %w", err)
	}

	return id, nil
}
