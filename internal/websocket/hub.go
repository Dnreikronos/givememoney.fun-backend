package websocket

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	Clients    map[uuid.UUID]map[*Client]bool
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

type Message struct {
	StreamerID uuid.UUID
	Data       interface{}
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID]map[*Client]bool),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Clients[client.StreamerID] == nil {
				h.Clients[client.StreamerID] = make(map[*Client]bool)
			}
			h.Clients[client.StreamerID][client] = true
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Clients[client.StreamerID]; ok {
				delete(clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.RLock()
			clients := h.Clients[message.StreamerID]
			h.mu.RUnlock()
			for client := range clients {
				select {
				case client.Send <- message.Data:
				default:
					close(client.Send)
					delete(h.Clients[message.StreamerID], client)
				}
			}
		}
	}
}

func (h *Hub) BroadcastToStreamer(streamerID uuid.UUID, data interface{}) {
	h.Broadcast <- &Message{StreamerID: streamerID, Data: data}
}
