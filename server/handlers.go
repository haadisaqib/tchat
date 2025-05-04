package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ====== Register Chatter ======

func registerChatter(scanner *bufio.Scanner, conn net.Conn) *Chatter {
	if scanner.Scan() {
		input := scanner.Text()
		fmt.Println("Received from client:", input)

		parts := strings.SplitN(input, "|", 4)
		if len(parts) != 4 {
			fmt.Fprintln(conn, "Invalid format. Expected: UUID|displayName|choice|roomData")
			return nil
		}

		uuidStr := strings.TrimSpace(parts[0])
		displayName := strings.TrimSpace(parts[1])
		choice := strings.TrimSpace(parts[2])
		roomData := strings.TrimSpace(parts[3])

		uuid, err := strconv.Atoi(uuidStr)
		if err != nil {
			fmt.Fprintln(conn, "Invalid UUID format. Must be a number.")
			return nil
		}

		chatter := newChatter(uuid, displayName, conn)
		if chatter == nil {
			return nil
		}

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
		sendMessage(chatter, room, message)
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
			deleteChatHistory(room.roomID)
			fmt.Printf("Room %d deleted\n", room.roomID)
		}
	}
	delete(server.chatters, chatter.UUID)
	fmt.Printf("Chatter %d deleted from server\n", chatter.UUID)
}
