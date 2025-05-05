package main

type Server struct {
	rooms    map[int]*Room
	chatters map[string]*Chatter 
}

var server = &Server{
	rooms:    make(map[int]*Room),
	chatters: make(map[string]*Chatter),
}
