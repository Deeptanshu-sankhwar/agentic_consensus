package handlers

import (
	"log"
	"net/http"

	"github.com/Deeptanshu-sankhwar/agentic_consensus/communication"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket manages WebSocket connections and broadcasts agent votes
func HandleWebSocket(c *gin.Context) {
	chainID := c.GetString("chainID")
	log.Printf("New WebSocket connection for chain: %s", chainID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	broadcast := func(data communication.AgentVote) {
		event := struct {
			Type    string                  `json:"type"`
			Payload communication.AgentVote `json:"payload"`
		}{
			Type:    "AGENT_VOTE",
			Payload: data,
		}

		log.Printf("Sending WebSocket event: %+v", event)
		err := conn.WriteJSON(event)
		if err != nil {
			log.Printf("Error writing to websocket: %v", err)
		}
	}

	log.Printf("Starting file watcher for chain: %s", chainID)
	go communication.WatchDiscussionFile(chainID, broadcast)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket connection closed: %v", err)
			break
		}
	}
}
