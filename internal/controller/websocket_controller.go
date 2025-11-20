package controller

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebsocketController struct {
	hub *websocket.Hub
}

func NewWebsocketController(hub *websocket.Hub) *WebsocketController {
	return &WebsocketController{
		hub: hub,
	}
}

func (wsc *WebsocketController) HandleConnection(ctx *gin.Context) {
	streamerID, err := uuid.Parse(ctx.Param("streamer_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid streamer_id"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := &websocket.Client{
		Hub:        wsc.hub,
		Conn:       conn,
		Send:       make(chan interface{}, 256),
		StreamerID: streamerID,
	}

	wsc.hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}
