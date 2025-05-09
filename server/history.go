package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func GetOrCreateChatHistory(roomID int) (string, error) {
	historyDir := "./rooms"
	filePath := filepath.Join(historyDir, fmt.Sprintf("%d.json", roomID))

	err := os.MkdirAll(historyDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to ensure directory %s exists: %w", historyDir, err)
	}

	_, err = os.Stat(filePath)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Room history file %s does not exist, creating...\n", filePath)
			file, createErr := os.Create(filePath)
			if createErr != nil {
				return "", fmt.Errorf("failed to create chat history file %s: %w", filePath, createErr)
			}
			file.Close()
			fmt.Printf("Room history file %s created successfully.\n", filePath)
			return filePath, nil
		} else {
			return "", fmt.Errorf("failed to check status of file %s: %w", filePath, err)
		}
	}

	fmt.Printf("Room history file %s already exists.\n", filePath)
	return filePath, nil
}

func deleteChatHistory(roomID int) {
	path := fmt.Sprintf("./rooms/%d.json", roomID)

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File does not exist — nothing to delete
		return
	} else if err != nil {
		// Some other error occurred when checking
		fmt.Println("Error checking file:", err)
		return
	}

	// Try to delete the file
	if err := os.Remove(path); err != nil {
		fmt.Println("Error deleting file:", err)
		return
	}

	fmt.Printf("Room %d deleted (file removed)\n", roomID)
}

// writeToJson appends a ChatMessage as a single JSON line to ./rooms/{roomID}.json
func writeToJson(roomID int, msg ChatMessage) error {
	path := fmt.Sprintf("./rooms/%d.json", roomID)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}
