package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type initPayload struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Choice      string `json:"choice"`
	RoomData    string `json:"roomData"`
}

type chatPayload struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsePayload struct {
	Type    string      `json:"type"`
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type errorPayload struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

var upgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer ws.Close()

	var hello initPayload
	if err := ws.ReadJSON(&hello); err != nil || hello.Type != "init" {
		log.Println("bad init:", err)
		return
	}
	if hello.ID == "" {
		hello.ID = uuid.New().String()
	}

	chatter := newWsChatter(hello.ID, hello.DisplayName, ws)
	if chatter == nil {
		_ = ws.WriteJSON(errorPayload{Type: "error", Message: "duplicate-uuid"})
		return
	}

	_ = incrementChatterCounter()

	var room *Room
	switch hello.Choice {
	case "1":
		capacity, _ := strconv.Atoi(hello.RoomData)
		if capacity < 1 || capacity > 20 {
			_ = ws.WriteJSON(errorPayload{Type: "error", Message: "invalid-capacity"})
			return
		}
		room = newRoom(capacity)
		server.rooms[room.roomID] = room
	case "2":
		rid, _ := strconv.Atoi(hello.RoomData)
		if !roomExists(rid) {
			_ = ws.WriteJSON(errorPayload{Type: "error", Message: "room-not-found"})
			return
		}
		existing := server.rooms[rid]
		if isRoomFull(existing) {
			_ = ws.WriteJSON(errorPayload{Type: "error", Message: "room-full"})
			return
		}
		room = existing
	default:
		_ = ws.WriteJSON(errorPayload{Type: "error", Message: "invalid-choice"})
		return
	}

	joinRoom(room, chatter)

	_ = ws.WriteJSON(responsePayload{
		Type:    "response",
		Event:   "joined",
		Payload: map[string]interface{}{"roomID": room.roomID},
	})

	if history, err := readHistory(room.roomID); err == nil {
		for _, cm := range history {
			_ = ws.WriteJSON(responsePayload{
				Type:  "response",
				Event: "history",
				Payload: map[string]string{
					"from": cm.Sender,
					"text": cm.Message,
				},
			})
		}
	}

	for {
		var msg chatPayload
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}
		if msg.Type == "message" && strings.TrimSpace(msg.Text) != "" {
			cm := ChatMessage{
				Sender:    chatter.DisplayName,
				Message:   msg.Text,
				Timestamp: time.Now().Format(time.RFC3339),
			}
			_ = writeToJson(room.roomID, cm)

			out := responsePayload{
				Type:  "response",
				Event: "message",
				Payload: map[string]string{
					"from": cm.Sender,
					"text": cm.Message,
				},
			}
			for _, c := range room.chatters {
				if c.WsConn != nil {
					_ = c.WsConn.WriteJSON(out)
				}
			}
		}
	}

	handleDisconnect(chatter, room)
}

func main() {
	// REST endpoint: GET /chatter-count
	http.HandleFunc("/chatter-count", func(w http.ResponseWriter, r *http.Request) {
		// Allow cross-origin requests for development
		w.Header().Set("Access-Control-Allow-Origin", "*") // Or limit to "http://localhost:5173"
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			// Handle preflight requests
			w.WriteHeader(http.StatusOK)
			return
		}

		cnt, err := getChatterCount()
		if err != nil {
			http.Error(w, "couldn't read counter", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"count": cnt})
	})

	// WebSocket endpoint
	http.HandleFunc("/ws", wsHandler)

	log.Println("[ws] listening on :9002")
	log.Fatal(http.ListenAndServe("127.0.0.1:9002", nil))
}
