// disconnect.go
package main

import (
	"fmt"
)


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