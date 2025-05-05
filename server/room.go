// room.go
package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
Room represents a single chat room.

capacity      – max simultaneous users (1‑20)
numOfChatter  – current user count
roomID        – 5‑digit random ID (unique per live server)
chatters      – slice of *Chatter currently inside
chatHistory   – path to the room’s JSON history file on disk
*/
type Room struct {
	capacity     int
	numOfChatter int
	roomID       int
	chatters     []*Chatter
	chatHistory  string
}

/* ------------------------------------------------------------------ */
/* Room helpers                                                       */
/* ------------------------------------------------------------------ */

// newRoom creates a room with a unique 5‑digit ID and prepares its
// on‑disk history file.
func newRoom(capacity int) *Room {
	rand.Seed(time.Now().UnixNano())

	for {
		id := rand.Intn(99999-10000) + 10000 // 10000‑99999
		if _, clash := server.rooms[id]; !clash {

			hFile, err := GetOrCreateChatHistory(id)
			if err != nil {
				fmt.Printf("Error creating history for room %d: %v\n", id, err)
				return nil
			}

			fmt.Printf("New room %d (cap=%d) created\n", id, capacity)
			return &Room{
				capacity:     capacity,
				numOfChatter: 0,
				roomID:       id,
				chatters:     []*Chatter{},
				chatHistory:  hFile,
			}
		}
	}
}

// joinRoom adds a chatter, prints / logs a summary, and dumps history to stdout.
// (The WebSocket handler sends the history back to the browser.)
func joinRoom(r *Room, ch *Chatter) {
	if isRoomFull(r) {
		fmt.Printf("Room %d is full (cap=%d)\n", r.roomID, r.capacity)
		return
	}

	r.numOfChatter++
	r.chatters = append(r.chatters, ch)
	ch.connectedTo = r.roomID

	hist := readChatHistory(r.roomID)
	if hist != "" {
		fmt.Printf("Loaded history for room %d (%d bytes)\n", r.roomID, len(hist))
	}

	fmt.Printf("Chatter %s joined room %d (%d/%d)\n",
		ch.UUID, r.roomID, r.numOfChatter, r.capacity)
}

func roomExists(roomID int) bool {
	_, exists := server.rooms[roomID]
	return exists
}

// isRoomFull returns true when the live population == capacity.
func isRoomFull(r *Room) bool {
	return r.numOfChatter >= r.capacity
}
