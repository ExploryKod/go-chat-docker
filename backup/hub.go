// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	ID uuid.UUID `json:"id"`

	Name string `json:"name"`

	Private bool `json:"private"`

	clients map[*Client]bool

	register chan *Client

	unregister chan *Client

	broadcast chan *Message
}

func newHub(name string, private bool) *Hub {
	return &Hub{
		ID:         uuid.New(),
		Name:       name,
		Private:    private,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

func (h *Hub) runHub() {
	for {
		select {

		case client := <-h.register:
			h.registerClientInHub(client)

		case client := <-h.unregister:
			h.unregisterClientInHub(client)

		case message := <-h.broadcast:
			h.broadcastToClientsInHub(message.encode())
		}
	}
}

func (h *Hub) registerClientInHub(client *Client) {
	if !h.Private {
		h.notifyClientJoined(client)
	}
	h.clients[client] = true
}

func (h *Hub) unregisterClientInHub(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
	}
}

func (h *Hub) broadcastToClientsInHub(message []byte) {
	for client := range h.clients {
		client.send <- message
	}
}

func (h *Hub) GetId() string {
	return h.ID.String()
}

func (h *Hub) GetName() string {
	return h.Name
}

const welcomeMessage = "%s a rejoint la salle !"

func (h *Hub) notifyClientJoined(client *Client) {
	message := &Message{
		Action:  SendMessageAction,
		Target:  h,
		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	}

	h.broadcastToClientsInHub(message.encode())
}

func (h *Hub) Hub(client *Client) {
	// by sending the message first the new user won't see his own message.
	h.notifyClientJoined(client)
	h.clients[client] = true
}
