// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websockets

import "encoding/json"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan broadcastRequest

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

type broadcastRequest struct {
	clientID string
	body     json.RawMessage
}

func NewHub() *Hub {
	h := &Hub{
		broadcast:  make(chan broadcastRequest),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}

	go h.run()

	return h
}

func (h *Hub) Broadcast(i string, b []byte) {
	h.broadcast <- broadcastRequest{
		clientID: i,
		body:     b,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case m := <-h.broadcast:
			for client := range h.clients {
				if client.identity.ClientID != m.clientID {
					continue
				}

				select {
				case client.send <- m.body:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
