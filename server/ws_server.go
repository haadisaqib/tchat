package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/* payloads from / to client */
type initPayload struct {
	Type        string `json:"type"` // must be "init"
	ID          string `json:"id"`   // browser‑generated UUID
	DisplayName string `json:"displayName"`
	Choice      string `json:"choice"`   // "1" create, "2" join
	RoomData    string `json:"roomData"` // capacity or roomID
}
type chatPayload struct{ Type, Text string }
type outgoing struct{ From, Text string }

/* upgrader */
var upgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

/* ---------------- WS handler ---------------- */

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer ws.Close()

	/* ---------- handshake ---------- */
	var hello initPayload
	if err := ws.ReadJSON(&hello); err != nil || hello.Type != "init" {
		log.Println("bad init:", err)
		return
	}
	if hello.ID == "" {
		hello.ID = uuid.New().String()
	}
	if _, dup := server.chatters[hello.ID]; dup {
		_ = ws.WriteJSON(outgoing{From: "system", Text: "duplicate‑uuid"})
		return
	}
	ch := &Chatter{UUID: hello.ID, DisplayName: hello.DisplayName, WsConn: ws}
	server.chatters[ch.UUID] = ch

	/* ---------- room create / join ---------- */
	var room *Room

	switch hello.Choice {
	case "1": // create
		cap, _ := strconv.Atoi(hello.RoomData)
		if cap < 1 || cap > 20 {
			_ = ws.WriteJSON(outgoing{From: "system", Text: "invalid‑capacity"})
			return
		}
		room = newRoom(cap)
		server.rooms[room.roomID] = room

		case "2": // JOIN an existing room
		rid, _ := strconv.Atoi(hello.RoomData)
	
		if !roomExists(rid) {                      // <─ NEW helper you added
			ws.WriteJSON(outgoing{From: "system", Text: "room-not-found"})
			return
		}
	
		rm := server.rooms[rid]
		if isRoomFull(rm) {
			ws.WriteJSON(outgoing{From: "system", Text: "room-full"})
			return
		}
	
		room = rm
	default:
		_ = ws.WriteJSON(outgoing{From: "system", Text: "invalid‑choice"})
		return
	}

	joinRoom(room, ch)

	/* ---------- confirmation + history ---------- */
	_ = ws.WriteJSON(outgoing{From: "system", Text: fmt.Sprintf("joined‑room %d", room.roomID)})

	if hist := readChatHistory(room.roomID); hist != "" {
		lines := strings.Split(strings.TrimSpace(hist), "\n")
		for _, ln := range lines {
			var cm ChatMessage
			if err := json.Unmarshal([]byte(ln), &cm); err == nil {
				_ = ws.WriteJSON(outgoing{From: cm.Sender, Text: cm.Message})
			}
		}
	}

	/* ---------- live chat loop ---------- */
	for {
		var msg chatPayload
		if err := ws.ReadJSON(&msg); err != nil { // client disconnect
			break
		}
		if msg.Type == "message" && strings.TrimSpace(msg.Text) != "" {
			broadcast(room, ch.DisplayName, msg.Text)
		}
	}

	handleDisconnect(ch, room)
}

/* broadcast to every chatter in the room */
func broadcast(room *Room, from, text string) {
	out := outgoing{From: from, Text: text}
	for _, c := range room.chatters {
		if c.WsConn != nil {
			_ = c.WsConn.WriteJSON(out)
		}
	}
	saveMessageToRoom(room.roomID, ChatMessage{Sender: from, Message: text, Timestamp: time.Now().Format(time.RFC3339)})
}

/* single entry‑point */
func main() {
	http.HandleFunc("/ws", wsHandler)
	log.Println("[ws] listening on :9002")
	log.Fatal(http.ListenAndServe(":9002", nil))
}
