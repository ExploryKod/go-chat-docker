// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Check if the origin is in the list of allowed origins
		allowedOrigins := []string{
			"https://chat-talks-client.vercel.app",
			"http://localhost:8002",
			"http://localhost:8003",
			"http://localhost:8081"}
		origin := r.Header.Get("Origin")
		fmt.Println("Request Origin:", origin)
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				return true
			} else {
				return true
			}
		}
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ID uuid.UUID `json:"id"`

	Name string `json:"name"`

	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	hubs map[*Hub]bool

	wsServer *WsServer
}

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	return &Client{
		ID:       uuid.New(),
		conn:     conn,
		send:     make(chan []byte, 256),
		hubs:     make(map[*Hub]bool),
		wsServer: wsServer,
		Name:     name,
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()
	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//client.hub.broadcast <- message
		client.handleNewMessage(jsonMessage)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 {
		log.Println("Url Param 'name' is missing")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := newClient(conn, wsServer, name[0])

	go client.writePump()
	go client.readPump()
	// new goroutines.
	// Allow collection of memory referenced by the caller by doing all work in
	wsServer.register <- client
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	for hub := range client.hubs {
		hub.unregister <- client
	}

	close(client.send)
	client.conn.Close()
}

func (client *Client) handleNewMessage(jsonMessage []byte) {

	var message Message

	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
	}

	// Attach the client object as the sender of the messsage.
	message.Sender = client

	switch message.Action {
	case SendMessageAction:
		hubName := message.Target.GetName()
		println("hubName: ", hubName)
		//if hub := client.wsServer.findHubByID(hubID); hub != nil {
		//	println("hub found")
		//	hub.broadcast <- &message
		//}
		if hub := client.wsServer.findHubByName(hubName); hub != nil {
			println("hub found")
			hub.broadcast <- &message
		}
	case JoinHubAction:
		client.handleJoinHubMessage(message)

	case LeaveHubAction:
		client.handleLeaveHubMessage(message)

	case JoinHubPrivateAction:
		client.handleJoinHubPrivateMessage(message)
	}
}

// client.go
func (client *Client) handleJoinHubMessage(message Message) {
	hubName := message.Message

	client.joinHub(hubName, nil)
}

func (client *Client) handleLeaveHubMessage(message Message) {
	hub := client.wsServer.findHubByID(message.Message)
	if hub == nil {
		return
	}
	if _, ok := client.hubs[hub]; ok {
		delete(client.hubs, hub)
	}

	hub.unregister <- client
}

// New method
// When joining a private room we will combine the IDs of the users
// Then we will bothe join the client and the target.
func (client *Client) handleJoinHubPrivateMessage(message Message) {

	target := client.wsServer.findClientByID(message.Message)
	if target == nil {
		return
	}

	// create unique room name combined to the two IDs
	hubName := message.Message + client.ID.String()

	client.joinHub(hubName, target)
	target.joinHub(hubName, client)

}

// New method
// Joining a room both for public and private roooms
// When joiing a private room a sender is passed as the opposing party
func (client *Client) joinHub(hubName string, sender *Client) {

	room := client.wsServer.findHubByName(hubName)
	if room == nil {
		room = client.wsServer.createHub(hubName, sender != nil)
	}

	// Don't allow to join private rooms through public room message
	if sender == nil && room.Private {
		return
	}

	if !client.isInRoom(room) {
		client.hubs[room] = true
		room.register <- client
		client.notifyRoomJoined(room, sender)
	}

}

// New method
// Check if the client is not yet in the room
func (client *Client) isInRoom(hub *Hub) bool {
	if _, ok := client.hubs[hub]; ok {
		return true
	}
	return false
}

// New method
// Notify the client of the new room he/she joined
func (client *Client) notifyRoomJoined(hub *Hub, sender *Client) {
	message := Message{
		Action:  HubJoinedAction,
		Target:  hub,
		Sender:  sender,
		Message: "a rejoins la salle",
	}

	client.send <- message.encode()
}

func (client *Client) GetName() string {
	return client.Name
}
