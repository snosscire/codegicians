package main

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_ttf"
)

type Texture struct {
	Texture *sdl.Texture
	Width   int32
	Height  int32
}

type TheCode struct {
	textures []*Texture
}

func NewTheCode(renderer *sdl.Renderer) *TheCode {
	font, err := ttf.OpenFont("data/font/Share-TechMono.ttf", 60)
	if err != nil {
		log.Fatalf("%v\n", err)
		return nil
	}
	txt, err := ioutil.ReadFile("data/map.txt")
	if err != nil {
		log.Fatalf("%v\n", err)
		return nil
	}
	lines := strings.Split(string(txt), "\n")
	color := sdl.Color{255, 255, 255, 128}
	textures := []*Texture{}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		surface, err := font.RenderUTF8_Blended(line, color)
		if err != nil {
			log.Fatalf("line:%s,err:%v\n", line, err)
		}
		texture, err := renderer.CreateTextureFromSurface(surface)
		if err != nil {
			log.Fatalf("line:%s,err:%v", line, err)
		}
		t := &Texture{
			texture,
			surface.W,
			surface.H,
		}
		textures = append(textures, t)
	}
	return &TheCode{
		textures: textures,
	}
}

func (tc *TheCode) Draw(renderer *sdl.Renderer, camera *Camera) {
	yOffset := int32(0)
	if camera.Y > 0 && (camera.Y+camera.H) < 1280 {
		for (camera.Y+camera.H+yOffset)%64 != 0 {
			yOffset += 1
		}
	}
	for i, t := range tc.textures {
		dst := sdl.Rect{0 - camera.X, (int32(i) * 64) - camera.Y + yOffset, t.Width, t.Height}
		renderer.Copy(t.Texture, nil, &dst)
	}
}
