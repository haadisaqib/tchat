// TCP Chatroom Server
package main

import (
	"bufio"
	"fmt"
	"net"
)

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
