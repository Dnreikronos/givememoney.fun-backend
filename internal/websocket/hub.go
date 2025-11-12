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

type Message struct {
	StreamerID uuid.UUID
	Data       interface{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.StreamerID] == nil {
				h.clients[client.StreamerID] = make(map[*Client]bool)
			}
			h.clients[client.StreamerID][client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.StreamerID]; ok {
				delete(clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.StreamerID]
			h.mu.RUnlock()
			for client := range clients {
				select {
				case client.send <- message.Data:
				default:
					close(client.send)
					delete(h.clients[message.StreamerID], client)
				}
			}
		}
	}
}

func (h *Hub) BroadcastToStreamer(streamerID uuid.UUID, data interface{}) {
	h.broadcast <- &Message{StreamerID: streamerID, Data: data}
}
