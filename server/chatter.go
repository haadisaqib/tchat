package main

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Chatter struct {
	UUID        string
	DisplayName string
	connectedTo int

	Conn   net.Conn
	WsConn *websocket.Conn
}

type ChatMessage struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// --- helpers -------------------------------------------------------------

func newWsChatter(id, name string, ws *websocket.Conn) *Chatter {
	// duplicate‐check
	if _, exists := server.chatters[id]; exists {
		return nil
	}
	// create & register
	c := &Chatter{
		UUID:        id,
		DisplayName: name,
		WsConn:      ws,
	}
	server.chatters[id] = c
	return c
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
