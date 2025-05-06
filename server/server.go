// server.go
package main

// Server holds all active rooms and chatters
type Server struct {
	rooms    map[int]*Room
	chatters map[string]*Chatter
}

// the one shared instance
var server = &Server{
	rooms:    make(map[int]*Room),
	chatters: make(map[string]*Chatter),
}
