package main

type WsServer struct {
	hubs    map[*Hub]bool
	clients map[*Client]bool

	unregister chan *Client
	register   chan *Client
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer() *WsServer {
	return &WsServer{
		hubs:       make(map[*Hub]bool),
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (server *WsServer) findHubByName(name string) *Hub {
	var foundHub *Hub
	for hub := range server.hubs {
		if hub.GetName() == name {
			foundHub = hub
			break
		}
	}

	return foundHub
}

func (server *WsServer) createHub(name string, private bool) *Hub {
	hub := newHub(name, private)
	go hub.runHub()
	server.hubs[hub] = true

	return hub
}

func (server *WsServer) Run() {
	for {
		select {

		case client := <-server.register:
			server.registerClient(client)

		case client := <-server.unregister:
			server.unregisterClient(client)
		}
	}
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		client.send <- message
	}
}

func (server *WsServer) registerClient(client *Client) {
	server.notifyClientJoined(client)
	server.listOnlineClients(client)
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
		server.notifyClientLeft(client)
	}
}

// chatServer.go
func (server *WsServer) notifyClientJoined(client *Client) {
	message := &Message{
		Action:  UserJoinedAction,
		Sender:  client,
		Message: client.Name + " a rejoint la salle",
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) notifyClientLeft(client *Client) {
	message := &Message{
		Action:  UserLeftAction,
		Sender:  client,
		Message: client.Name + " a quitté la salle",
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) listOnlineClients(client *Client) {
	for existingClient := range server.clients {
		message := &Message{
			Action:  UserJoinedAction,
			Sender:  existingClient,
			Message: existingClient.Name + " a rejoint la salle",
		}
		client.send <- message.encode()
	}
}

func (server *WsServer) findHubByID(ID string) *Hub {
	var foundHub *Hub
	for hub := range server.hubs {
		if hub.GetId() == ID {
			foundHub = hub
			break
		}
	}

	return foundHub
}

func (server *WsServer) findClientByID(ID string) *Client {
	var foundClient *Client
	for client := range server.clients {
		if client.ID.String() == ID {
			foundClient = client
			break
		}
	}

	return foundClient
}
