package main

const (
	MESSAGE_GAME_START        = '1'
	MESSAGE_GAME_END          = '3'
	MESSAGE_PLAYER_TELEPORT   = 't'
	MESSAGE_PLAYER_DAMAGE     = 'a'
	MESSAGE_PLAYER_RESPAWN    = 's'
	MESSAGE_PLAYER_DISCONNECT = '2'
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

type MessagePlayerDamage struct {
	Amount int
}

type MessagePlayerRespawn struct {
	X float32
	Y float32
}
