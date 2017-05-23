package main

const (
	MESSAGE_GAME_START = '1'
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
