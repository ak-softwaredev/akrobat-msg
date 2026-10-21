package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clientsMux *sync.RWMutex
	clients    map[int64]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
}

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	chatID int64
}

type Event struct {
	ChatID int64
	Data   []byte
}
