// TCP Chatroom Server
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

// ====== Data Structures ======

type Server struct {
	rooms    map[int]*Room
	chatters map[int]*Chatter
}

type Chatter struct {
	UUID         int
	DisplayName  string
	connectedTo  int
	tempChoice   string
	tempRoomData string
}

type Room struct {
	capacity     int
	numOfChatter int
	roomID       int
	messages     []string
	chatters     []*Chatter
	chatHistory  string
}

var server = &Server{
	rooms:    make(map[int]*Room),
	chatters: make(map[int]*Chatter),
}

// ====== Connection Handler ======

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New connection from:", conn.RemoteAddr())
	scanner := bufio.NewScanner(conn)

	chatter := registerChatter(scanner, conn)
	if chatter == nil {
		return
	}

	defer handleDisconnect(chatter, nil)

	room := handleJoinOrCreate(scanner, chatter, conn)
	if room == nil {
		return
	}

	handleChatLoop(scanner, chatter, room, conn)
	handleDisconnect(chatter, room)
}

// ====== Register Chatter ======

func registerChatter(scanner *bufio.Scanner, conn net.Conn) *Chatter {
	if scanner.Scan() {
		input := scanner.Text()
		parts := strings.SplitN(input, "|", 3)
		if len(parts) != 3 {
			fmt.Fprintln(conn, "Invalid format. Expected: displayName|choice|roomData")
			return nil
		}

		displayName := strings.TrimSpace(parts[0])
		choice := strings.TrimSpace(parts[1])
		roomData := strings.TrimSpace(parts[2])

		chatter := newChatter(displayName)
		server.chatters[chatter.UUID] = chatter

		chatter.tempChoice = choice
		chatter.tempRoomData = roomData

		fmt.Printf("Registered: %s (UUID %d)\n", chatter.DisplayName, chatter.UUID)
		return chatter
	}
	return nil
}

// ====== Handle Join / Create Room ======

func handleJoinOrCreate(scanner *bufio.Scanner, chatter *Chatter, conn net.Conn) *Room {
	choice := chatter.tempChoice
	roomData := chatter.tempRoomData

	if choice == "1" {
		capacity, err := strconv.Atoi(roomData)
		if err != nil || capacity < 1 || capacity > 20 {
			fmt.Fprintln(conn, "Invalid room capacity. Must be 1–20.")
			return nil
		}
		room := newRoom(capacity)
		server.rooms[room.roomID] = room
		joinRoom(room, chatter)
		fmt.Fprintf(conn, "Room created with ID %d.\n", room.roomID)
		return room

	} else if choice == "2" {
		for {
			roomID, err := strconv.Atoi(roomData)
			if err != nil {
				fmt.Fprintln(conn, "Invalid room ID. Please enter a number:")
			} else {
				room, exists := server.rooms[roomID]
				if exists {
					if isRoomFull(room) {
						fmt.Fprintln(conn, "Room is full. Try another room:")
					} else {
						joinRoom(room, chatter)
						fmt.Fprintf(conn, "You have joined room %d.\n", roomID)
						return room
					}
				} else {
					fmt.Fprintf(conn, "Room %d does not exist. Try again:\n", roomID)
				}
			}

			// Prompt for new input
			fmt.Fprint(conn, "> ")
			if !scanner.Scan() {
				return nil
			}
			roomData = strings.TrimSpace(scanner.Text())
		}
	}

	fmt.Fprintln(conn, "Invalid choice.")
	return nil
}

// ====== Chat Loop ======

func handleChatLoop(scanner *bufio.Scanner, chatter *Chatter, room *Room, conn net.Conn) {
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Println("Message from", chatter.DisplayName, ":", message)
		broadcast(room, fmt.Sprintf("%s: %s", chatter.DisplayName, message), chatter)
	}
}

func broadcast(room *Room, msg string, sender *Chatter) {
	room.messages = append(room.messages, msg)
	for _, c := range room.chatters {
		if c.UUID != sender.UUID {
			// Normally you'd map UUID to conn; this version assumes broadcast placeholder
			fmt.Printf("→ To %d: %s\n", c.UUID, msg)
		}
	}
}

// ====== Disconnect Cleanup ======

func handleDisconnect(chatter *Chatter, room *Room) {
	if room != nil && chatter.connectedTo == room.roomID {
		room.numOfChatter--
		newList := []*Chatter{}
		for _, c := range room.chatters {
			if c.UUID != chatter.UUID {
				newList = append(newList, c)
			}
		}
		room.chatters = newList
		fmt.Printf("Chatter %d left room %d\n", chatter.UUID, room.roomID)
		if room.numOfChatter == 0 {
			delete(server.rooms, room.roomID)
			fmt.Printf("Room %d deleted\n", room.roomID)
		}
	}
	delete(server.chatters, chatter.UUID)
	fmt.Printf("Chatter %d deleted from server\n", chatter.UUID)
}

// ====== Helpers ======

func newRoom(capacity int) *Room {
	rand.Seed(time.Now().UnixNano())
	for {
		roomID := rand.Intn(99999-10000) + 10000
		if _, exists := server.rooms[roomID]; !exists {
			//print
			fmt.Printf("New room created with ID %d\n", roomID)
			//create json TODO
			createChatHistory(roomID)
			return &Room{capacity: capacity, roomID: roomID, chatters: []*Chatter{}}
		}
	}
}

func createChatHistory(roomID int) string {
	//Check if roomID.json exists in  ./rooms
	//If it does not exist, create it
	return fmt.Sprintf("./rooms/%d.json", roomID)
}

func joinRoom(room *Room, chatter *Chatter) {
	if isRoomFull(room) {
		fmt.Println("Room is full")
		return
	}
	room.numOfChatter++
	room.chatters = append(room.chatters, chatter)
	chatter.connectedTo = room.roomID
	fmt.Printf("Chatter %d joined room %d\n", chatter.UUID, room.roomID)
}

func isRoomFull(room *Room) bool {
	return room.numOfChatter >= room.capacity
}

func newChatter(name string) *Chatter {
	rand.Seed(time.Now().UnixNano())
	for {
		uuid := rand.Intn(999999-100000) + 100000
		if _, exists := server.chatters[uuid]; !exists {
			return &Chatter{UUID: uuid, DisplayName: name}
		}
	}
}

// ====== Main Server ======

func main() {
	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Server listening on port 9001...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
}
