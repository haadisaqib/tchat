package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)


type Chatter struct {
	UUID        string
	DisplayName string
	connectedTo int

	// ↓ add these two lines back in
	tempChoice   string
	tempRoomData string

	Conn   net.Conn
	WsConn *websocket.Conn
}

type ChatMessage struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// --- helpers -------------------------------------------------------------

func newChatter(uuid, name string, conn net.Conn) *Chatter {
	if _, exists := server.chatters[uuid]; exists {
		if conn != nil {
			fmt.Fprintln(conn, "Duplicate UUID detected. You are already connected.")
		}
		return nil
	}
	return &Chatter{UUID: uuid, DisplayName: name, Conn: conn}
}

// Fallback UUID for non‑browser callers (rare)
func UUIDgenerator(addr string) string {
	ip := addr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	ip = strings.Trim(ip, "[]")
	ipNoDots := strings.ReplaceAll(ip, ".", "")
	n, err := strconv.Atoi(ipNoDots)
	if err != nil {
		return strconv.Itoa(int(time.Now().UnixNano()))
	}
	if n%2 == 0 {
		n += 10
	} else {
		n += 11
	}
	return strconv.Itoa(n)
}

// --- message persistence/broadcast --------------------------------------

func saveMessageToRoom(roomID int, msg ChatMessage) {
	path := fmt.Sprintf("./rooms/%d.json", roomID)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open chat history:", err)
		return
	}
	defer f.Close()

	b, _ := json.Marshal(msg)
	f.Write(b)
	f.Write([]byte("\n"))
}

// Legacy TCP helper (still useful in tests)
func sendMessage(ch *Chatter, room *Room, text string) {
	msg := ChatMessage{Sender: ch.DisplayName, Message: text, Timestamp: time.Now().Format(time.RFC3339)}
	saveMessageToRoom(room.roomID, msg)

	for _, c := range room.chatters {
		if c.UUID == ch.UUID {
			continue
		}
		if c.WsConn != nil {
			_ = c.WsConn.WriteJSON(struct{ From, Text string }{From: msg.Sender, Text: msg.Message})
		} else if c.Conn != nil {
			fmt.Fprintln(c.Conn, msg.Sender+": "+msg.Message)
		}
	}
}
