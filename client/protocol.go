package main

type NetworkMessage byte

const (
	MESSAGE_GAME_START        NetworkMessage = '1'
	MESSAGE_PLAYER_MOVE_UP    NetworkMessage = 'u'
	MESSAGE_PLAYER_MOVE_DOWN  NetworkMessage = 'd'
	MESSAGE_PLAYER_MOVE_LEFT  NetworkMessage = 'l'
	MESSAGE_PLAYER_MOVE_RIGHT NetworkMessage = 'r'
)
