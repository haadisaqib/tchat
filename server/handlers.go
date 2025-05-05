// handlers.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

/*
This file exists **only** for legacy CLI / telnet clients.
The browser frontend talks through WebSockets and never hits this code.

If you never plan on supporting plain‑TCP clients again, feel free
to delete handlers.go and the TCP portions of the project.
*/

// registerChatter reads a single line:
//
//	UUID|displayName|choice|roomData
//
// and boots the connection if the format is wrong.
func registerChatter(sc *bufio.Scanner, conn net.Conn) *Chatter {
	if !sc.Scan() {
		return nil
	}
	line := sc.Text()
	parts := strings.SplitN(line, "|", 4)
	if len(parts) != 4 {
		fmt.Fprintln(conn, "Invalid format. Expected: UUID|displayName|choice|roomData")
		return nil
	}

	uuid := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	ch := newChatter(uuid, name, conn)
	if ch == nil {
		return nil // duplicate UUID message already sent
	}

	// Preserve user’s intention so the CLI flow works
	ch.tempChoice = strings.TrimSpace(parts[2])   // "1" create, "2" join
	ch.tempRoomData = strings.TrimSpace(parts[3]) // capacity or roomID

	server.chatters[ch.UUID] = ch
	fmt.Printf("Registered CLI chatter %s (%s)\n", ch.DisplayName, ch.UUID)
	return ch
}

func handleDisconnect(ch *Chatter, room *Room) {
	if room != nil && ch.connectedTo == room.roomID {
		room.numOfChatter--
		// rebuild active‑chatter slice
		newList := []*Chatter{}
		for _, c := range room.chatters {
			if c.UUID != ch.UUID {
				newList = append(newList, c)
			}
		}
		room.chatters = newList

		fmt.Printf("Chatter %s left room %d\n", ch.UUID, room.roomID)

		if room.numOfChatter == 0 {
			delete(server.rooms, room.roomID)
			deleteChatHistory(room.roomID)
			fmt.Printf("Room %d deleted (empty)\n", room.roomID)
		}
	}
	delete(server.chatters, ch.UUID)
	fmt.Printf("Chatter %s removed from server\n", ch.UUID)
}
