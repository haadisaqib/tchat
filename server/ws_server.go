// ws_server.go
package main

import (
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
	if _, dup := server.chatters[hello.ID]; dup {
		_ = ws.WriteJSON(errorPayload{Type: "error", Message: "duplicate-uuid"})
		return
	}
	chatter := &Chatter{UUID: hello.ID, DisplayName: hello.DisplayName, WsConn: ws}
	server.chatters[chatter.UUID] = chatter

	var room *Room
	switch hello.Choice {
	case "1":
		cap, _ := strconv.Atoi(hello.RoomData)
		if cap < 1 || cap > 20 {
			_ = ws.WriteJSON(errorPayload{Type: "error", Message: "invalid-capacity"})
			return
		}
		room = newRoom(cap)
		server.rooms[room.roomID] = room

	case "2":
		rid, _ := strconv.Atoi(hello.RoomData)
		if !roomExists(rid) {
			_ = ws.WriteJSON(errorPayload{Type: "error", Message: "room-not-found"})
			return
		}
		rm := server.rooms[rid]
		if isRoomFull(rm) {
			_ = ws.WriteJSON(errorPayload{Type: "error", Message: "room-full"})
			return
		}
		room = rm

	default:
		_ = ws.WriteJSON(errorPayload{Type: "error", Message: "invalid-choice"})
		return
	}

	joinRoom(room, chatter)

	// send joined confirmation
	_ = ws.WriteJSON(responsePayload{
		Type:    "response",
		Event:   "joined",
		Payload: map[string]interface{}{"roomID": room.roomID},
	})

	// send chat history
	history, err := readHistory(room.roomID)
	if err == nil {
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

	// live chat loop
	for {
		var msg chatPayload
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}
		if msg.Type == "message" && strings.TrimSpace(msg.Text) != "" {
			// persist
			cm := ChatMessage{Sender: chatter.DisplayName, Message: msg.Text, Timestamp: time.Now().Format(time.RFC3339)}
			_ = writeToJson(room.roomID, cm)

			// broadcast
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
	http.HandleFunc("/ws", wsHandler)
	log.Println("[ws] listening on :9002")
	log.Fatal(http.ListenAndServe(":9002", nil))
}
