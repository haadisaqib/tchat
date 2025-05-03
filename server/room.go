package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Room struct {
	capacity     int
	numOfChatter int
	roomID       int
	chatters     []*Chatter
	chatHistory  string
}

func newRoom(capacity int) *Room {
	rand.Seed(time.Now().UnixNano())
	for {
		roomID := rand.Intn(99999-10000) + 10000
		if _, exists := server.rooms[roomID]; !exists {
			//print
			fmt.Printf("New room created with ID %d\n", roomID)
			//create json TODO
			chatHistory, err := GetOrCreateChatHistory(roomID)
			if err != nil {
				fmt.Printf("Error creating chat history: %v\n", err)
				return nil
			}
			fmt.Printf("Chat history file route is %s\n", chatHistory)
			return &Room{capacity: capacity, roomID: roomID, chatters: []*Chatter{}, chatHistory: chatHistory}
		}
	}
}

func joinRoom(room *Room, chatter *Chatter) {
	//check if room full
	if isRoomFull(room) {
		fmt.Println("Room is full")
		return
	}
	room.numOfChatter++
	room.chatters = append(room.chatters, chatter)
	chatter.connectedTo = room.roomID

	history := readChatHistory(room.roomID)

	if history != "" {
		fmt.Printf("Chat history for room %d:\n%s\n", room.roomID, history)
	} else {
		fmt.Printf("No chat history found for room %d\n", room.roomID)
	}

	fmt.Printf("Chatter %d joined room %d\n", chatter.UUID, room.roomID)
}

func isRoomFull(room *Room) bool {
	return room.numOfChatter >= room.capacity
}
