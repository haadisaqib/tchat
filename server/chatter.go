package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

type Chatter struct {
	UUID         int
	DisplayName  string
	connectedTo  int
	tempChoice   string
	tempRoomData string
	Conn         net.Conn
}

type ChatMessage struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func newChatter(name string, conn net.Conn) *Chatter {
	rand.Seed(time.Now().UnixNano())
	for {
		uuid := rand.Intn(999999-100000) + 100000
		if _, exists := server.chatters[uuid]; !exists {
			return &Chatter{UUID: uuid, DisplayName: name, connectedTo: 0, Conn: conn}
		}
	}
}

func saveMessageToRoom(roomID int, msg ChatMessage) {
	path := fmt.Sprintf("./rooms/%d.json", roomID)

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open chat history:", err)
		return
	}
	defer file.Close()

	jsonBytes, _ := json.Marshal(msg)
	file.Write(jsonBytes)
	file.Write([]byte("\n"))
}

func sendMessage(chatter *Chatter, room *Room, messageText string) {
	msg := ChatMessage{
		Sender:    chatter.DisplayName,
		Message:   messageText,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	saveMessageToRoom(room.roomID, msg)

	for _, c := range room.chatters {
		if c.UUID != chatter.UUID {
			formatted := fmt.Sprintf("%s: %s", msg.Sender, msg.Message)
			fmt.Fprintln(c.Conn, formatted)
		}
	}
}
