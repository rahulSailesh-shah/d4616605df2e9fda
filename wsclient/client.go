package wsclient

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func Connect(url string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Send(msg []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *Client) Receive() ([]byte, error) {
	_, data, err := c.conn.ReadMessage()
	return data, err
}

func (c *Client) Listen(onMessage func([]byte), onDone func(error)) {
	go func() {
		for {
			msg, err := c.Receive()
			if err != nil {
				if onDone != nil {
					onDone(err)
				}
				return
			}
			if onMessage != nil {
				onMessage(msg)
			}
		}
	}()
}

func (c *Client) Close() error {
	return c.conn.Close()
}
