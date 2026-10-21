package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 16 * 1024
)

type Hub struct {
	clients    map[int64]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
}

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	chatID int64
	userID int64
	hub    *Hub
}

type Event struct {
	ChatID int64
	Data   []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.chatID] == nil {
				h.clients[client.chatID] = make(map[*Client]bool)
			}

			h.clients[client.chatID][client] = true

		case client := <-h.unregister:
			h.removeClient(client)

		case event := <-h.broadcast:
			for client := range h.clients[event.ChatID] {
				select {
				case client.send <- event.Data:
					// nothing, hooray!
				default:
					h.removeClient(client)
				}
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	clients, ok := h.clients[client.chatID]
	if !ok {
		return
	}

	if _, exists := clients[client]; !exists {
		return
	}

	delete(clients, client)
	close(client.send)

	if len(clients) == 0 {
		delete(h.clients, client.chatID)
	}
}

func (h *Hub) Publish(chatID int64, data []byte) {
	h.broadcast <- Event{
		ChatID: chatID,
		Data:   data,
	}
}
