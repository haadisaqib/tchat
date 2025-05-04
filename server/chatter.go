package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

func newChatter(uuid int, name string, conn net.Conn) *Chatter {
	if _, exists := server.chatters[uuid]; exists {
		fmt.Fprintf(conn, "Duplicate UUID detected. You are already connected.\n")
		return nil
	}
	return &Chatter{
		UUID:        uuid,
		DisplayName: name,
		connectedTo: 0,
		Conn:        conn,
	}
}

func UUIDgenerator(conn net.Conn) int {
	ip := conn.RemoteAddr()
	ipString := ip.String()
	lastColon := strings.LastIndex(ipString, ":")
	if lastColon != -1 {
		ipString = ipString[:lastColon]
	}
	ipString = strings.Trim(ipString, "[]")
	ipWithoutPeriods := strings.ReplaceAll(ipString, ".", "")
	// Convert the ipwithout periods to an integer
	ipInt, err := strconv.Atoi(ipWithoutPeriods)
	if err != nil {
		fmt.Println("Error converting IP to integer:", err)
		return 0
	}
	// Check if the integer is even or odd
	if ipInt%2 == 0 {
		ipInt += 10
	} else {
		ipInt += 11
	}
	return ipInt
}

// example:
//127.0.0.1 = 127001 % 2 == 1 therefore + 11 = 127012
//198.1.168.254 = 1981168254 % 2 == 0 therefore + 10 = 1981168264
//10.10.10 = 101010 % 2 == 0 therefore + 10 = 101020

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
