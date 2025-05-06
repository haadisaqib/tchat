package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

// readHistory returns all ChatMessages from ./rooms/{roomID}.json
func readHistory(roomID int) ([]ChatMessage, error) {
	path := fmt.Sprintf("./rooms/%d.json", roomID)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var msgs []ChatMessage
	for scanner.Scan() {
		var cm ChatMessage
		if err := json.Unmarshal(scanner.Bytes(), &cm); err == nil {
			msgs = append(msgs, cm)
		}
	}
	return msgs, scanner.Err()
}

func readChatHistory(roomID int) string {
	//Check if roomID.json exists in  ./rooms
	file, err := os.Open("./rooms")

	if err != nil {
		fmt.Println("Error opening directory:", err)
		return ""
	}
	defer file.Close()

	//Check if roomID.json exists
	files, err := file.Readdirnames(0)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return ""
	}
	for _, f := range files {
		if f == fmt.Sprintf("%d.json", roomID) {
			data, err := os.ReadFile(fmt.Sprintf("./rooms/%d.json", roomID))
			if err != nil {
				fmt.Println("Error reading file:", err)
				return ""
			}
			return string(data)
		}
	}
	return ""
}

// historyHandler streams the ./rooms/{roomID}.json file as a JSON array.
func historyHandler(w http.ResponseWriter, r *http.Request) {
	// Expects: /history?roomId=12345
	q := r.URL.Query().Get("roomId")
	id, err := strconv.Atoi(q)
	if err != nil {
		http.Error(w, "invalid roomId", http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("./rooms/%d.json", id)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "history not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var msgs []ChatMessage
	for scanner.Scan() {
		var cm ChatMessage
		if err := json.Unmarshal(scanner.Bytes(), &cm); err == nil {
			msgs = append(msgs, cm)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

func init() {
	// register the HTTP endpoint alongside your WebSocket
	http.HandleFunc("/history", historyHandler)
}
