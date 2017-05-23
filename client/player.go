package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
)

const (
	PLAYER1_TEXTURE_PATH     string  = "data/player1.png"
	PLAYER2_TEXTURE_PATH     string  = "data/player2.png"
	PLAYER_WIDTH             int32   = 64
	PLAYER_HEIGHT            int32   = 64
	PLAYER_TELEPORT_COOLDOWN float32 = 1000.0
	PLAYER_TELEPORT_SPEED    float32 = 0.5
)

type Position struct {
	X float32
	Y float32
}

type PlayerDirection struct {
	Up    bool
	Down  bool
	Left  bool
	Right bool
}

type Player struct {
	me          bool
	Position    Position
	Direction   PlayerDirection
	texture     *sdl.Texture
	drawTexture bool
	health      int
	dying       bool

	teleporting      bool
	teleportCooldown float32
	teleportPosition Position
	teleportAlpha    float32
	teleportRect     sdl.Rect
	teleportRectW    float32
	teleportRectH    float32
}

func NewPlayer(renderer *sdl.Renderer, me bool, texturePath string) *Player {
	texture, err := img.LoadTexture(renderer, texturePath)
	if err != nil {
		panic(err)
	}
	player := &Player{
		me:          me,
		health:      100,
		texture:     texture,
		drawTexture: true,
	}
	return player
}

func (p *Player) IsTeleporting() bool {
	return p.teleporting
}

func (p *Player) screenPosition(camera *Camera) (int32, int32) {
	var x int32 = int32(p.Position.X) - camera.X - (PLAYER_WIDTH / 2)
	var y int32 = int32(p.Position.Y) - camera.Y - (PLAYER_HEIGHT / 2)
	if p.me {
		if camera.X > 0 && (camera.X+camera.W) < 1280 {
			x = (SCREEN_WIDTH / 2) - (PLAYER_WIDTH / 2)
		}
		if camera.Y > 0 && (camera.Y+camera.H) < 1280 {
			y = (SCREEN_HEIGHT / 2) - (PLAYER_HEIGHT / 2)
		}
	}
	return x, y
}

func (p *Player) updateTeleportRect(camera *Camera) {
	x, y := p.screenPosition(camera)
	p.teleportRect.X = x + (PLAYER_WIDTH / 2) - (p.teleportRect.W / 2)
	p.teleportRect.Y = y + (PLAYER_HEIGHT / 2) - (p.teleportRect.H / 2)
	p.teleportRect.W = int32(p.teleportRectW)
	p.teleportRect.H = int32(p.teleportRectH)
}

func (p *Player) Teleport(x, y float32) bool {
	if p.teleporting || p.teleportCooldown > 0.0 || !p.IsAlive() {
		return false
	}
	if x < 0 || x > 1280 || y < 0 || y > 1280 {
		return false
	}
	p.teleporting = true
	p.teleportCooldown = PLAYER_TELEPORT_COOLDOWN
	p.teleportPosition.X = x
	p.teleportPosition.Y = y
	p.teleportRectW = 1.0
	p.teleportRectH = 1.0
	p.teleportAlpha = 255.0
	return true
}

func (p *Player) IsAlive() bool {
	return p.health > 0
}

func (p *Player) TakeDamage(damage int) {
	p.health -= damage
	if p.health <= 0 {
		p.dying = true
		p.teleportRectW = 1.0
		p.teleportRectH = 1.0
		p.teleportAlpha = 255.0
	}
}

func (p *Player) Kill() {
	p.TakeDamage(100)
}

func (p *Player) Update(deltaTime float32) {
	if p.teleporting || p.dying {
		if p.teleportAlpha <= 0.0 {
			if p.teleporting {
				if p.Position.X != p.teleportPosition.X || p.Position.Y != p.teleportPosition.Y {
					p.Position.X = p.teleportPosition.X
					p.Position.Y = p.teleportPosition.Y
					p.teleportRectW = 1.0
					p.teleportRectH = 1.0
					p.teleportAlpha = 255.0
				} else {
					p.teleporting = false
				}
			}
		} else if p.teleportRect.W >= PLAYER_WIDTH {
			p.drawTexture = false
			if p.teleporting && p.Position.X == p.teleportPosition.X && p.Position.Y == p.teleportPosition.Y {
				p.drawTexture = true
			}
			p.teleportAlpha -= deltaTime * PLAYER_TELEPORT_SPEED
		} else {
			increase := deltaTime * PLAYER_TELEPORT_SPEED
			p.teleportRectW += increase
			p.teleportRectH += increase
		}
	}
	if p.teleportCooldown > 0.0 {
		p.teleportCooldown -= deltaTime
	}
}

func (p *Player) Draw(renderer *sdl.Renderer, camera *Camera) {
	x, y := p.screenPosition(camera)
	if p.drawTexture {
		rect := sdl.Rect{
			x,
			y,
			PLAYER_WIDTH,
			PLAYER_HEIGHT,
		}
		renderer.Copy(p.texture, nil, &rect)
	}
	if p.teleporting || p.dying {
		p.updateTeleportRect(camera)
		renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
		renderer.SetDrawColor(255, 255, 255, uint8(p.teleportAlpha))
		renderer.FillRect(&p.teleportRect)
	}
}
