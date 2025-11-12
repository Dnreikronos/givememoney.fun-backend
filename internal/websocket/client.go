package websocket

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan interface{}
	StreamerID uuid.UUID
}
