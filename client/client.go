package main

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net"
)

type Client struct {
	connection           net.Conn
	connectionReadWriter *bufio.ReadWriter
	messageDecoder       *gob.Decoder
	messageEncoder       *gob.Encoder
	disconnectHandler    func()
	messageHandler       func(NetworkMessage, interface{})
}

func (c *Client) SetDisconnectHandler(handler func()) {
	c.disconnectHandler = handler
}

func (c *Client) SetMessageHandler(handler func(NetworkMessage, interface{})) {
	c.messageHandler = handler
}

func (c *Client) handleDisconnect() {
	if c.disconnectHandler != nil {
		c.disconnectHandler()
	}
}

func (c *Client) handleMessage(msg byte, data interface{}) {
	log.Printf("Command: %d\n", msg)
	if c.messageHandler != nil {
		c.messageHandler(NetworkMessage(msg), data)
	}
}

func (c *Client) Read() {
	defer c.connection.Close()
	for {
		msg, err := c.connectionReadWriter.ReadByte()
		switch {
		case err == io.EOF:
			c.handleDisconnect()
			return
		case err != nil:
			c.handleDisconnect()
			return
		}
		switch NetworkMessage(msg) {
		case MESSAGE_GAME_START:
			var data MessageGameStart
			err := c.messageDecoder.Decode(&data)
			if err != nil {
				log.Printf("%v\n", err)
				continue
			}
			log.Printf("%v\n", data)
			c.handleMessage(msg, data)
		case MESSAGE_PLAYER_TELEPORT:
			var data MessagePlayerTeleport
			err := c.messageDecoder.Decode(&data)
			if err != nil {
				log.Printf("%v\n", err)
				continue
			}
			log.Printf("%v\n", data)
			c.handleMessage(msg, &data)
		default:
			c.handleMessage(msg, nil)
		}
	}
}

func (c *Client) send(msg byte, data interface{}) {
	err := c.connectionReadWriter.WriteByte(msg)
	if err != nil {
		return
	}
	if data != nil {
		err = c.messageEncoder.Encode(data)
		if err != nil {
			return
		}
	}
	err = c.connectionReadWriter.Flush()
}

func (c *Client) Send(msg NetworkMessage, data interface{}) {
	go c.send(byte(msg), data)
}
