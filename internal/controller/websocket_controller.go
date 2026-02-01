package controller

import (
	"net/http"

	apperrors "github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WebsocketController handles WebSocket connections for real-time updates.
type WebsocketController struct {
	hub *websocket.Hub
}

// NewWebsocketController creates a new WebsocketController with the given hub.
func NewWebsocketController(hub *websocket.Hub) *WebsocketController {
	return &WebsocketController{
		hub: hub,
	}
}

// HandleConnection upgrades the HTTP connection to WebSocket and registers the client.
func (wsc *WebsocketController) HandleConnection(ctx *gin.Context) {
	streamerID, err := uuid.Parse(ctx.Param("streamer_id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid streamer_id", err))
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
