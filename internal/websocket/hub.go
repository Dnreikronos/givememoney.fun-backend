package websocket

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

