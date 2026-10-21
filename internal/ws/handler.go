package ws

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *Hub) HandleChat(c *gin.Context) {
	userID := c.GetInt64("userID")

	chatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || chatID <= 0 || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid ChatID"},
		})
		return
	}

	conn, err := upgrader.Upgrade(
		c.Writer,
		c.Request,
		nil,
	)
	if err != nil {
		log.Printf("[ERROR] websocket upgrade: %v", err)
		return
	}

	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		chatID: chatID,
		userID: userID,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}
