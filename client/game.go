package main

import (
	"bufio"
	"encoding/gob"
	"log"
	"net"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
)

const (
	DefaultPort      = ":46337"
	SCREEN_WIDTH     = 1280
	SCREEN_HEIGHT    = 720
	PATH_TEXTURE_MAP = "data/map.png"
)

type GameState int

const (
	STATE_MAINMENU   GameState = 0
	STATE_CONNECTING GameState = 1
	STATE_STARTING   GameState = 2
	STATE_PLAYING    GameState = 3
)

type Camera sdl.Rect

func (c *Camera) Update(p *Player) {
	c.X = int32(p.Position.X) + (PLAYER_WIDTH / 2) - (SCREEN_WIDTH / 2)
	c.Y = int32(p.Position.Y) + (PLAYER_HEIGHT / 2) - (SCREEN_HEIGHT / 2)

	if c.X < 0 {
		c.X = 0
	} else if c.X > (1280 - SCREEN_WIDTH) {
		c.X = 1280 - SCREEN_WIDTH
	}

	if c.Y < 0 {
		c.Y = 0
	} else if c.Y > (1280 - SCREEN_HEIGHT) {
		c.Y = 1280 - SCREEN_HEIGHT
	}

	c.W = SCREEN_WIDTH
	c.H = SCREEN_HEIGHT
}

type Game struct {
	client       *Client
	window       *sdl.Window
	renderer     *sdl.Renderer
	running      bool
	state        GameState
	localPlayer  *Player
	otherPlayer  *Player
	mapTexture   *sdl.Texture
	camera       Camera
	startMessage MessageGameStart
}

func NewGame() *Game {
	game := &Game{
		running: false,
		state:   STATE_MAINMENU,
	}
	return game
}

func (g *Game) handleMessage(msg NetworkMessage, data interface{}) {
	switch msg {
	case MESSAGE_GAME_START:
		log.Println("Start game")
		g.startMessage = data.(MessageGameStart)
		g.state = STATE_PLAYING
	case MESSAGE_PLAYER_MOVE_UP:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			y -= float32(PLAYER_HEIGHT)
			g.otherPlayer.Teleport(x, y)
		}
	case MESSAGE_PLAYER_MOVE_DOWN:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			y += float32(PLAYER_HEIGHT)
			g.otherPlayer.Teleport(x, y)
		}
	case MESSAGE_PLAYER_MOVE_LEFT:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			x -= float32(PLAYER_WIDTH)
			g.otherPlayer.Teleport(x, y)
		}
	case MESSAGE_PLAYER_MOVE_RIGHT:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			x += float32(PLAYER_WIDTH)
			g.otherPlayer.Teleport(x, y)
		}
	}
}

func (g *Game) handleKeyDown(event *sdl.KeyDownEvent) {
	if g.state == STATE_PLAYING && g.localPlayer != nil {
		if event.Keysym.Sym == sdl.K_F12 {
			g.localPlayer.Kill()
			return
		}
		if g.localPlayer.IsTeleporting() {
			return
		}
		x := g.localPlayer.Position.X
		y := g.localPlayer.Position.Y
		switch event.Keysym.Sym {
		case sdl.K_UP, sdl.K_k:
			if !g.localPlayer.Direction.Up {
				g.localPlayer.Direction.Up = true
				y -= float32(PLAYER_HEIGHT)
				if g.localPlayer.Teleport(x, y) {
					g.client.Send(MESSAGE_PLAYER_MOVE_UP, nil)
				}
			}
		case sdl.K_DOWN, sdl.K_j:
			if !g.localPlayer.Direction.Down {
				g.localPlayer.Direction.Down = true
				y += float32(PLAYER_HEIGHT)
				if g.localPlayer.Teleport(x, y) {
					g.client.Send(MESSAGE_PLAYER_MOVE_DOWN, nil)
				}
			}
		case sdl.K_LEFT, sdl.K_h:
			if !g.localPlayer.Direction.Left {
				g.localPlayer.Direction.Left = true
				x -= float32(PLAYER_WIDTH)
				if g.localPlayer.Teleport(x, y) {
					g.client.Send(MESSAGE_PLAYER_MOVE_LEFT, nil)
				}
			}
		case sdl.K_RIGHT, sdl.K_l:
			if !g.localPlayer.Direction.Right {
				g.localPlayer.Direction.Right = true
				x += float32(PLAYER_WIDTH)
				if g.localPlayer.Teleport(x, y) {
					g.client.Send(MESSAGE_PLAYER_MOVE_RIGHT, nil)
				}
			}
		}
	}
}

func (g *Game) handleKeyUp(event *sdl.KeyUpEvent) {
	if g.state == STATE_PLAYING && g.localPlayer != nil {
		switch event.Keysym.Sym {
		case sdl.K_UP, sdl.K_k:
			g.localPlayer.Direction.Up = false
		case sdl.K_DOWN, sdl.K_j:
			g.localPlayer.Direction.Down = false
		case sdl.K_LEFT, sdl.K_h:
			g.localPlayer.Direction.Left = false
		case sdl.K_RIGHT, sdl.K_l:
			g.localPlayer.Direction.Right = false
		}
	}
}

func (g *Game) handleInput() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			g.running = false
		case *sdl.KeyDownEvent:
			g.handleKeyDown(e)
		case *sdl.KeyUpEvent:
			g.handleKeyUp(e)
		}
	}
}

func (g *Game) run() {
	if g.running {
		return
	}
	g.running = true

	sdl.Init(sdl.INIT_EVERYTHING)

	var err error
	var windowFlags uint32 /*= sdl.WINDOW_FULLSCREEN_DESKTOP*/

	g.window, err = sdl.CreateWindow("Codegicians",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		1280, 720,
		windowFlags)
	if err != nil {
		panic(err)
	}
	defer g.window.Destroy()

	g.renderer, err = sdl.CreateRenderer(g.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer g.renderer.Destroy()

	//g.renderer.SetLogicalSize(SCREEN_WIDTH, SCREEN_HEIGHT)

	currentTime := sdl.GetTicks()
	lastTime := currentTime
	var deltaTime float32

	for g.running {
		currentTime = sdl.GetTicks()
		deltaTime = float32(currentTime - lastTime)
		lastTime = currentTime

		g.handleInput()

		if g.state == STATE_PLAYING {
			if g.localPlayer != nil {
				g.localPlayer.Update(deltaTime)
				g.camera.Update(g.localPlayer)
			}
			if g.otherPlayer != nil {
				g.otherPlayer.Update(deltaTime)
			}
		}

		g.renderer.SetDrawColor(0, 0, 0, 255)
		g.renderer.Clear()

		if g.state == STATE_PLAYING {
			if g.mapTexture == nil {
				g.mapTexture, _ = img.LoadTexture(g.renderer, PATH_TEXTURE_MAP)
			}
			if g.mapTexture != nil {
				mapDstRect := sdl.Rect{0, 0, SCREEN_WIDTH, SCREEN_HEIGHT}
				mapSrcRect := sdl.Rect(g.camera)
				g.renderer.Copy(g.mapTexture, &mapSrcRect, &mapDstRect)
			}
			if g.localPlayer == nil {
				g.localPlayer = NewPlayer(g.renderer, true, g.startMessage.MyTexture)
				g.localPlayer.Position.X = g.startMessage.MyPosX
				g.localPlayer.Position.Y = g.startMessage.MyPosY

			}
			if g.otherPlayer == nil {
				g.otherPlayer = NewPlayer(g.renderer, false, g.startMessage.EnemyTexture)
				g.otherPlayer.Position.X = g.startMessage.EnemyPosX
				g.otherPlayer.Position.Y = g.startMessage.EnemyPosY
			}

			if g.otherPlayer != nil {
				g.otherPlayer.Draw(g.renderer, &g.camera)
			}
			if g.localPlayer != nil {
				g.localPlayer.Draw(g.renderer, &g.camera)
			}
		}

		g.renderer.Present()
	}
}

func (g *Game) Connect(address string) {
	g.state = STATE_CONNECTING
	connection, err := net.Dial("tcp", address+DefaultPort)
	if err != nil {
		return
	}
	readWriter := bufio.NewReadWriter(bufio.NewReader(connection), bufio.NewWriter(connection))
	g.client = &Client{
		connection:           connection,
		connectionReadWriter: readWriter,
		messageDecoder:       gob.NewDecoder(readWriter),
		messageEncoder:       gob.NewEncoder(readWriter),
	}
	g.client.SetMessageHandler(g.handleMessage)
	go g.client.Read()
	g.state = STATE_STARTING
	g.run()
}

func (g *Game) MainMenu() {
	g.state = STATE_MAINMENU
	g.run()
}
