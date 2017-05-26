package main

type NetworkMessage byte

const (
	MESSAGE_GAME_START        NetworkMessage = '1'
	MESSAGE_GAME_END          NetworkMessage = '3'
	MESSAGE_PLAYER_MOVE_UP    NetworkMessage = 'u'
	MESSAGE_PLAYER_MOVE_DOWN  NetworkMessage = 'd'
	MESSAGE_PLAYER_MOVE_LEFT  NetworkMessage = 'l'
	MESSAGE_PLAYER_MOVE_RIGHT NetworkMessage = 'r'
	MESSAGE_PLAYER_TELEPORT   NetworkMessage = 't'
	MESSAGE_PLAYER_DAMAGE     NetworkMessage = 'a'
	MESSAGE_PLAYER_DIE        NetworkMessage = 'k'
	MESSAGE_PLAYER_RESPAWN    NetworkMessage = 's'
	MESSAGE_PLAYER_DISCONNECT NetworkMessage = '2'
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
