package main

import (
	"log"
)

type Game struct {
	players []*Player
}

func NewGame(clients []*Client) *Game {
	game := new(Game)
	for _, client := range clients {
		client.SetDisconnectHandler(game.handlePlayerDisconnect)
		client.SetMessageHandler(game.handlePlayerMessage)
		player := NewPlayer(client)
		game.players = append(game.players, player)
	}
	return game
}

func (g *Game) handlePlayerDisconnect(client *Client) {
	log.Printf("(Game) Player disconnected.\n")
}

func (g *Game) sendToAllExcept(msg byte, data interface{}, client *Client) {
	for _, player := range g.players {
		if player.ClientId() != client.Id() {
			player.Send(msg, data)
		}
	}
}

func (g *Game) handlePlayerMessage(client *Client, msg byte, data interface{}) {
	log.Printf("(Game) Received player message: %d\n", msg)
	switch msg {
	case 'u', 'd', 'l', 'r':
		g.sendToAllExcept(msg, data, client)
	}
}

func (g *Game) Start() {
	for _, player := range g.players {
		player.Send(MESSAGE_GAME_START, nil)
	}
	//for {
	//}
}
