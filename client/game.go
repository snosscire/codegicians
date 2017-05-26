package main

import (
	"bufio"
	"encoding/gob"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
	"github.com/veandco/go-sdl2/sdl_ttf"
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

type GameMode bool

const (
	MODE_INSERT  GameMode = true
	MODE_COMMAND GameMode = false
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

type Target interface {
	TakeDamage(amount int, client *Client)
	ScreenPosition(camera *Camera) (int32, int32)
	IsAlive() bool
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
	startMessage *MessageGameStart
	theCode      *TheCode
	showTheCode  bool
	gKeyPressed  bool
	nKeyPressed  string
	mode         GameMode

	targetWords                     []string
	insertModeFont                  *ttf.Font
	currentWord                     string
	currentWordTexture              *sdl.Texture
	currentWordTextureWidth         int32
	currentWordTextureHeight        int32
	currentTarget                   Target
	currentTargetWords              []string
	currentTargetWordsTexture       *sdl.Texture
	currentTargetWordsTextureWidth  int32
	currentTargetWordsTextureHeight int32
}

func NewGame() *Game {
	game := &Game{
		running: false,
		state:   STATE_MAINMENU,
	}
	return game
}

func (g *Game) findYPosInCode(y int32) int32 {
	offset := int32(0)
	for (y+offset)%64 != 0 {
		offset += 1
	}
	return y + offset
}

func (g *Game) handleNavigationCommands(event *sdl.KeyDownEvent) bool {
	match := false
	switch event.Keysym.Sym {
	case sdl.K_0:
		if len(g.nKeyPressed) == 0 { // jump to the start of the line
			g.localPlayer.Teleport(32, g.localPlayer.Position.Y)
			g.setTarget(nil)
			match = true
		} else { // go to line n
			g.nKeyPressed += "0"
			g.gKeyPressed = false
			return true
		}
	case sdl.K_1, sdl.K_2, sdl.K_3, sdl.K_5, sdl.K_6, sdl.K_7, sdl.K_8, sdl.K_9:
		g.nKeyPressed += string(event.Keysym.Sym)
		g.gKeyPressed = false
		return true
	case sdl.K_DOLLAR, sdl.K_4: // jump to the end of the line
		if event.Keysym.Sym == sdl.K_DOLLAR || event.Keysym.Mod&sdl.KMOD_LSHIFT > 0 || event.Keysym.Mod&sdl.KMOD_RALT > 0 {
			g.localPlayer.Teleport(1280-32, g.localPlayer.Position.Y)
			g.setTarget(nil)
			match = true
		} else {
			g.nKeyPressed += "4"
			g.gKeyPressed = false
			return true
		}
	case sdl.K_h: // move to top of screen
		if event.Keysym.Mod&sdl.KMOD_LSHIFT > 0 || event.Keysym.Mod&sdl.KMOD_RSHIFT > 0 {
			g.localPlayer.Teleport(g.localPlayer.Position.X, float32(g.findYPosInCode(g.camera.Y)+32))
			g.setTarget(nil)
			match = true
		}
	case sdl.K_l: // move to bottom of screen
		if event.Keysym.Mod&sdl.KMOD_LSHIFT > 0 || event.Keysym.Mod&sdl.KMOD_RSHIFT > 0 {
			g.localPlayer.Teleport(g.localPlayer.Position.X, float32(g.findYPosInCode(g.camera.Y+g.camera.H)-32))
			g.setTarget(nil)
			match = true
		}
	case sdl.K_m: // move to middle of screen
		if event.Keysym.Mod&sdl.KMOD_LSHIFT > 0 || event.Keysym.Mod&sdl.KMOD_RSHIFT > 0 {
			y := g.camera.Y + (g.camera.H / 2)
			g.localPlayer.Teleport(g.localPlayer.Position.X, float32(g.findYPosInCode(y)-32))
			g.setTarget(nil)
			match = true
		}
	case sdl.K_g:
		if event.Keysym.Mod&sdl.KMOD_LSHIFT > 0 || event.Keysym.Mod&sdl.KMOD_RSHIFT > 0 {
			if len(g.nKeyPressed) == 0 { // go to the last line of the document
				g.localPlayer.Teleport(g.localPlayer.Position.X, float32(1280-32))
				g.setTarget(nil)
				match = true
			} else { // go to line n
				line, _ := strconv.Atoi(g.nKeyPressed)
				if line >= 1 && line <= 20 {
					line--
					g.localPlayer.Teleport(g.localPlayer.Position.X, float32((line*64)+32))
					g.setTarget(nil)
					g.nKeyPressed = ""
					match = true
				}
			}
		} else {
			if g.gKeyPressed { // go to the first line of the document
				g.localPlayer.Teleport(g.localPlayer.Position.X, float32(32))
				g.setTarget(nil)
				g.gKeyPressed = false
				match = true
			} else {
				g.gKeyPressed = true
				g.nKeyPressed = ""
				return true
			}
		}
	case sdl.K_b:
		if g.theCode != nil {
			x := g.theCode.PreviousWordAtBeginningMapPosition(g.localPlayer.Position.X, g.localPlayer.Position.Y)
			g.localPlayer.Teleport(x, g.localPlayer.Position.Y)
			g.setTarget(nil)
		}
		match = true
	case sdl.K_e:
		if g.theCode != nil {
			x := g.theCode.NextWordAtEndMapPosition(g.localPlayer.Position.X, g.localPlayer.Position.Y)
			g.localPlayer.Teleport(x, g.localPlayer.Position.Y)
			g.setTarget(nil)
		}
		match = true
	case sdl.K_w:
		if g.theCode != nil {
			x := g.theCode.NextWordAtBeginningMapPosition(g.localPlayer.Position.X, g.localPlayer.Position.Y)
			g.localPlayer.Teleport(x, g.localPlayer.Position.Y)
			g.setTarget(nil)
		}
		match = true
	case sdl.K_LSHIFT, sdl.K_RSHIFT:
		return false
	}
	if match {
		teleportMsg := MessagePlayerTeleport{
			g.localPlayer.TeleportPosition.X,
			g.localPlayer.TeleportPosition.Y,
		}
		g.client.Send(MESSAGE_PLAYER_TELEPORT, &teleportMsg)
		g.gKeyPressed = false
		g.nKeyPressed = ""
		return true
	}
	return false
}

func (g *Game) randomDamageAmount() int {
	return rand.Intn(10) + 10
}

func (g *Game) randomTargetWord() string {
	if len(g.targetWords) == 0 {
		txt, err := ioutil.ReadFile("data/words.txt")
		if err != nil {
			log.Fatalf("%v\n", err)
			return ""
		}
		g.targetWords = strings.Split(string(txt), "\n")
	}
	random := rand.Intn(len(g.targetWords) - 1)
	return g.targetWords[random]
}

func (g *Game) setTarget(target Target) {
	g.currentWord = ""
	g.currentTarget = target
	if target == nil {
		g.mode = MODE_COMMAND
	}
	g.updateCurrentTargetWords()
}

func (g *Game) updateCurrentWordTexture() {
	//if len(g.currentWord) == 0 {
	//if g.currentWordTexture != nil {
	//g.currentWordTexture.Destroy()
	//g.currentWordTexture = nil
	//}
	//return
	//}
	color := sdl.Color{255, 255, 255, 255}
	surface, err := g.insertModeFont.RenderUTF8_Blended(g.currentWord+"_", color)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	if g.currentWordTexture != nil {
		g.currentWordTexture.Destroy()
	}
	g.currentWordTexture, err = g.renderer.CreateTextureFromSurface(surface)
	surface.Free()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	g.currentWordTextureWidth = surface.W
	g.currentWordTextureHeight = surface.H
}

func (g *Game) updateCurrentTargetWords() {
	if g.currentTarget == nil {
		if g.currentTargetWordsTexture != nil {
			g.currentTargetWordsTexture.Destroy()
			g.currentTargetWordsTexture = nil
		}
		g.currentTargetWords = []string{}
		return
	}
	for len(g.currentTargetWords) < 5 {
		g.currentTargetWords = append(g.currentTargetWords, g.randomTargetWord())
	}
	targetWords := strings.Join(g.currentTargetWords, " ")

	color := sdl.Color{255, 255, 255, 255}
	surface, err := g.insertModeFont.RenderUTF8_Blended(targetWords, color)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	if g.currentTargetWordsTexture != nil {
		g.currentTargetWordsTexture.Destroy()
	}
	g.currentTargetWordsTexture, err = g.renderer.CreateTextureFromSurface(surface)
	surface.Free()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	g.currentTargetWordsTextureWidth = surface.W
	g.currentTargetWordsTextureHeight = surface.H
}

func (g *Game) handleLocalPlayerDie() {
	g.setTarget(nil)
}

func (g *Game) handleInsertMode(event *sdl.KeyDownEvent) {
	if event.Keysym.Sym == sdl.K_BACKSPACE {
		if len(g.currentWord) > 0 {
			index := len(g.currentWord) - 1
			g.currentWord = g.currentWord[:index]
			g.updateCurrentWordTexture()
		}
	} else {
		key := int(event.Keysym.Sym)
		if key >= 97 && key <= 122 {
			g.currentWord += string(key)
			g.updateCurrentWordTexture()
		}
	}
	currentWord := g.currentWord
	if len(currentWord) > 0 && g.currentTarget != nil && len(g.currentTargetWords) > 0 {
		if currentWord == g.currentTargetWords[0] {
			g.currentTarget.TakeDamage(g.randomDamageAmount(), g.client)
			g.currentWord = ""
			newList := []string{}
			for i, word := range g.currentTargetWords {
				if i > 0 {
					newList = append(newList, word)
				}
			}
			g.currentTargetWords = newList
			g.updateCurrentWordTexture()
			g.updateCurrentTargetWords()
		}
	}
}

func (g *Game) handleKeyDown(event *sdl.KeyDownEvent) {
	if g.state == STATE_PLAYING && g.localPlayer != nil {
		currentMode := g.mode
		if currentMode == MODE_INSERT {
			if event.Keysym.Sym == sdl.K_ESCAPE {
				g.mode = MODE_COMMAND
			} else {
				g.handleInsertMode(event)
			}
			return
		} else if currentMode == MODE_COMMAND {
			if event.Keysym.Sym == sdl.K_i || event.Keysym.Sym == sdl.K_INSERT {
				g.mode = MODE_INSERT
				return
			}
		}
		if event.Keysym.Sym == sdl.K_F12 {
			g.localPlayer.TakeDamage(100, g.client)
			return
		} else if event.Keysym.Sym == sdl.K_F1 {
			g.showTheCode = !g.showTheCode
			return
		} else if event.Keysym.Sym == sdl.K_n {
			if g.otherPlayer.IsAlive() && int32(g.otherPlayer.Position.X) > g.camera.X && int32(g.otherPlayer.Position.X) < (g.camera.X+g.camera.W) && int32(g.otherPlayer.Position.Y) > g.camera.Y && int32(g.otherPlayer.Position.Y) < (g.camera.Y+g.camera.H) {
				g.setTarget(g.otherPlayer)
				log.Printf("targeted other player")
			} else {
				g.setTarget(nil)
				log.Printf("no target")
			}
		} else {
			if g.handleNavigationCommands(event) {
				return
			}
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
					g.setTarget(nil)
				}
			}
		case sdl.K_DOWN, sdl.K_j:
			if !g.localPlayer.Direction.Down {
				g.localPlayer.Direction.Down = true
				y += float32(PLAYER_HEIGHT)
				if g.localPlayer.Teleport(x, y) {
					g.client.Send(MESSAGE_PLAYER_MOVE_DOWN, nil)
					g.setTarget(nil)
				}
			}
		case sdl.K_LEFT, sdl.K_h:
			if !g.localPlayer.Direction.Left {
				g.localPlayer.Direction.Left = true
				x -= float32(PLAYER_WIDTH)
				if g.localPlayer.Teleport(x, y) {
					g.client.Send(MESSAGE_PLAYER_MOVE_LEFT, nil)
					g.setTarget(nil)
				}
			}
		case sdl.K_RIGHT, sdl.K_l:
			if !g.localPlayer.Direction.Right {
				g.localPlayer.Direction.Right = true
				x += float32(PLAYER_WIDTH)
				if g.localPlayer.Teleport(x, y) {
					g.client.Send(MESSAGE_PLAYER_MOVE_RIGHT, nil)
					g.setTarget(nil)
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

func (g *Game) handleUserEvent(event *sdl.UserEvent) {
	switch NetworkMessage(event.Code) {
	case MESSAGE_GAME_START:
		log.Println("Event: Start game")
		g.startMessage = (*MessageGameStart)(event.Data1)
		g.state = STATE_PLAYING
	case MESSAGE_PLAYER_TELEPORT:
		g.setTarget(nil)
		teleportMsg := (*MessagePlayerTeleport)(event.Data1)
		g.otherPlayer.Teleport(teleportMsg.X, teleportMsg.Y)
	case MESSAGE_PLAYER_DAMAGE:
		damageMsg := (*MessagePlayerDamage)(event.Data1)
		g.localPlayer.TakeDamage(damageMsg.Amount, g.client)
	case MESSAGE_PLAYER_DIE:
		g.setTarget(nil)
		g.otherPlayer.Die()
	case MESSAGE_PLAYER_RESPAWN:
		respawnMsg := (*MessagePlayerRespawn)(event.Data1)
		g.otherPlayer.Respawn(respawnMsg.X, respawnMsg.Y)
	case MESSAGE_PLAYER_MOVE_UP:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			y -= float32(PLAYER_HEIGHT)
			g.otherPlayer.Teleport(x, y)
			g.setTarget(nil)
		}
	case MESSAGE_PLAYER_MOVE_DOWN:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			y += float32(PLAYER_HEIGHT)
			g.otherPlayer.Teleport(x, y)
			g.setTarget(nil)
		}
	case MESSAGE_PLAYER_MOVE_LEFT:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			x -= float32(PLAYER_WIDTH)
			g.otherPlayer.Teleport(x, y)
			g.setTarget(nil)
		}
	case MESSAGE_PLAYER_MOVE_RIGHT:
		if g.state == STATE_PLAYING && g.otherPlayer != nil {
			x := g.otherPlayer.Position.X
			y := g.otherPlayer.Position.Y
			x += float32(PLAYER_WIDTH)
			g.otherPlayer.Teleport(x, y)
			g.setTarget(nil)
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
		case *sdl.UserEvent:
			g.handleUserEvent(e)
		}
	}
}

func (g *Game) run() {
	if g.running {
		return
	}
	g.running = true

	rand.Seed(time.Now().UTC().UnixNano())
	sdl.Init(sdl.INIT_EVERYTHING)
	ttf.Init()

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

	g.insertModeFont, err = ttf.OpenFont("data/font/Share-TechMono.ttf", 16)
	if err != nil {
		panic(err)
	}

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
			if g.currentWordTexture == nil {
				g.updateCurrentWordTexture()
			}
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
				g.localPlayer.StartPosition.X = g.startMessage.MyPosX
				g.localPlayer.StartPosition.Y = g.startMessage.MyPosY
				g.localPlayer.OnPlayerDie = g.handleLocalPlayerDie
			}
			if g.otherPlayer == nil {
				g.otherPlayer = NewPlayer(g.renderer, false, g.startMessage.EnemyTexture)
				g.otherPlayer.Position.X = g.startMessage.EnemyPosX
				g.otherPlayer.Position.Y = g.startMessage.EnemyPosY
				g.otherPlayer.StartPosition.X = g.startMessage.MyPosX
				g.otherPlayer.StartPosition.Y = g.startMessage.MyPosY
			}

			if g.currentTarget != nil && g.currentTarget.IsAlive() {
				tx, ty := g.currentTarget.ScreenPosition(&g.camera)
				g.renderer.SetDrawColor(255, 0, 0, 128)
				g.renderer.FillRect(&sdl.Rect{
					X: tx,
					Y: ty,
					W: 64,
					H: 64,
				})
			}
			if g.otherPlayer != nil {
				g.otherPlayer.Draw(g.renderer, &g.camera)
			}
			if g.localPlayer != nil {
				g.localPlayer.Draw(g.renderer, &g.camera)
			}
			if g.theCode == nil {
				g.theCode = NewTheCode(g.renderer)
			}
			if g.theCode != nil && g.showTheCode {
				g.theCode.Draw(g.renderer, &g.camera)
			}
			if g.mode == MODE_INSERT {
				g.drawInsertMode()
			}
		}

		g.renderer.Present()
	}

	ttf.Quit()
	sdl.Quit()
}

func (g *Game) drawInsertMode() {
	bgRect := sdl.Rect{0, SCREEN_HEIGHT - 40, SCREEN_WIDTH, 40}
	g.renderer.SetDrawColor(0, 0, 0, 255)
	g.renderer.FillRect(&bgRect)
	if g.currentTargetWordsTexture != nil {
		twRect := sdl.Rect{
			0,
			SCREEN_HEIGHT - g.currentWordTextureHeight - g.currentTargetWordsTextureHeight,
			g.currentTargetWordsTextureWidth,
			g.currentTargetWordsTextureHeight,
		}
		g.renderer.Copy(g.currentTargetWordsTexture, nil, &twRect)
	}
	if g.currentWordTexture != nil {
		cwRect := sdl.Rect{
			0,
			SCREEN_HEIGHT - g.currentWordTextureHeight,
			g.currentWordTextureWidth,
			g.currentWordTextureHeight,
		}
		g.renderer.Copy(g.currentWordTexture, nil, &cwRect)
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
	go g.client.Read()
	g.state = STATE_STARTING
	g.run()
}

func (g *Game) MainMenu() {
	g.state = STATE_MAINMENU
	g.run()
}
