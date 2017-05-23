package main

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net"
)

type Client struct {
	id                   int
	connection           net.Conn
	connectionReadWriter *bufio.ReadWriter
	messageDecoder       *gob.Decoder
	messageEncoder       *gob.Encoder
	disconnectHandler    func(*Client)
	messageHandler       func(*Client, byte, interface{})
}

func NewClient(conn net.Conn, id int) *Client {
	client := new(Client)
	client.id = id
	client.connection = conn
	client.connectionReadWriter = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	client.messageDecoder = gob.NewDecoder(conn)
	client.messageEncoder = gob.NewEncoder(client.connectionReadWriter)
	return client
}

func (c *Client) Id() int {
	return c.id
}

func (c *Client) SetDisconnectHandler(handler func(*Client)) {
	c.disconnectHandler = handler
}

func (c *Client) SetMessageHandler(handler func(*Client, byte, interface{})) {
	c.messageHandler = handler
}

func (c *Client) Disconnect() {
	c.connection.Close()
}

func (c *Client) handleDisconnect() {
	if c.disconnectHandler != nil {
		c.disconnectHandler(c)
	}
}

func (c *Client) handleMessage(msg byte) {
	log.Printf("Command: %d\n", msg)
	var data interface{}
	if c.messageHandler != nil {
		c.messageHandler(c, msg, data)
	}
}

func (c *Client) Read() {
	defer c.Disconnect()
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
		c.handleMessage(msg)
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

func (c *Client) Send(msg byte, data interface{}) {
	go c.send(msg, data)
}