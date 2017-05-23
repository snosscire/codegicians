package main

import (
	"log"
	"net"
	"sync"
)

type Server struct {
	networkListener     net.Listener
	nextClientId        int
	clientsWaiting      []*Client
	clientsWaitingMutex *sync.Mutex
}

func NewServer() *Server {
	server := new(Server)
	server.clientsWaitingMutex = new(sync.Mutex)
	return server
}

func (s *Server) StartNewGame(clients []*Client) {
	log.Println("StartNewGame()")
	game := NewGame(clients)
	go game.Start()
}

func (s *Server) handleWaitingClientDisconnect(disconnectedClient *Client) {
	s.clientsWaitingMutex.Lock()
	log.Printf("Client disconnected.\n")
	newList := make([]*Client, 0)
	for _, client := range s.clientsWaiting {
		if client.Id() != disconnectedClient.Id() {
			newList = append(newList, client)
		}
	}
	s.clientsWaiting = newList
	s.clientsWaitingMutex.Unlock()
}

func (s *Server) Run() {
	var err error
	s.networkListener, err = net.Listen("tcp", ":46337")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	for {
		conn, err := s.networkListener.Accept()
		if err != nil {
			log.Printf("%v", err)
			continue
		}

		s.nextClientId++
		client := NewClient(conn, s.nextClientId)
		client.SetDisconnectHandler(s.handleWaitingClientDisconnect)

		s.clientsWaitingMutex.Lock()
		s.clientsWaiting = append(s.clientsWaiting, client)

		if len(s.clientsWaiting) == 2 {
			s.StartNewGame(s.clientsWaiting)
			s.clientsWaiting = nil
		}
		s.clientsWaitingMutex.Unlock()

		go client.Read()
	}

}
