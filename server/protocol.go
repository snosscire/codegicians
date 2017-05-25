package main

const (
	MESSAGE_GAME_START      = '1'
	MESSAGE_PLAYER_TELEPORT = 't'
)

type MessageGameStart struct {
	MyClientId    int
	MyTexture     string
	MyPosX        float32
	MyPosY        float32
	EnemyClientId int
	EnemyTexture  string
	EnemyPosX     float32
	EnemyPosY     float32
}

type MessagePlayerTeleport struct {
	X float32
	Y float32
}
