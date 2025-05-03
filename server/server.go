package main

type Server struct {
	rooms    map[int]*Room
	chatters map[int]*Chatter
}

var server = &Server{
	rooms:    make(map[int]*Room),
	chatters: make(map[int]*Chatter),
}
