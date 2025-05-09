// disconnect.go
package main

import (
	"fmt"
)
// handleDisconnect removes a chatter from the server and updates the room
// occupancy. If the room is empty, it deletes the room and its history file.
// It also broadcasts the updated occupancy to all remaining chatters in the room.
func handleDisconnect(ch *Chatter, room *Room) {
	if room != nil && ch.connectedTo == room.roomID {
		room.numOfChatter--
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
		} else {
			broadcastRoomOccupancy(room)
		}
	}
	delete(server.chatters, ch.UUID)
	fmt.Printf("Chatter %s removed from server\n", ch.UUID)
}