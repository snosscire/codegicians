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
	g.sendToAllExcept(MESSAGE_PLAYER_DISCONNECT, client)
}

func (g *Game) sendToAllExcept(msg byte, client *Client) {
	for _, player := range g.players {
		if player.ClientId() != client.Id() {
			player.Send(msg)
		}
	}
}

func (g *Game) sendDataToAllExcept(msg byte, data interface{}, client *Client) {
	for _, player := range g.players {
		if player.ClientId() != client.Id() {
			player.SendData(msg, data)
		}
	}
}

func (g *Game) handlePlayerMessage(client *Client, msg byte, data interface{}) {
	log.Printf("(Game) Received player message: %s\n", string(msg))
	switch msg {
	case 'u', 'd', 'l', 'r', 'k', '3':
		g.sendToAllExcept(msg, client)
	case 't', 'a', 's':
		g.sendDataToAllExcept(msg, data, client)
	}
}

func (g *Game) Start() {
	for i, player := range g.players {
		var myPosX float32 = 1280.0 - 32.0
		var myPosY float32 = 32.0
		var myTexture string = "data/player1.png"
		var enemyPosX float32 = 32.0
		var enemyPosY float32 = 1280.0 - 32.0
		var enemyTexture string = "data/player2.png"
		if i > 0 {
			myPosX = 32.0
			myPosY = 1280.0 - 32.0
			myTexture = "data/player2.png"
			enemyPosX = 1280.0 - 32.0
			enemyPosY = 32.0
			enemyTexture = "data/player1.png"
		}
		data := MessageGameStart{
			MyClientId:   player.ClientId(),
			MyPosX:       myPosX,
			MyPosY:       myPosY,
			MyTexture:    myTexture,
			EnemyPosX:    enemyPosX,
			EnemyPosY:    enemyPosY,
			EnemyTexture: enemyTexture,
		}
		player.SendData(MESSAGE_GAME_START, &data)
	}
	//for {
	//}
}
