package main

type Player struct {
	client *Client
}

func NewPlayer(client *Client) *Player {
	player := new(Player)
	player.client = client
	return player
}

func (p *Player) ClientId() int {
	return p.client.Id()
}

func (p *Player) Send(msg byte) {
	p.client.Send(msg)
}

func (p *Player) SendData(msg byte, data interface{}) {
	p.client.SendData(msg, data)
}
