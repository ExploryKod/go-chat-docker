package main

import (
	"encoding/json"
	"log"
)

const SendMessageAction = "send-message"
const JoinHubAction = "join-hub"
const LeaveHubAction = "leave-hub"
const UserJoinedAction = "user-join"
const UserLeftAction = "user-left"
const JoinHubPrivateAction = "join-hub-private"
const HubJoinedAction = "hub-joined"

type Message struct {
	Action  string  `json:"action"`
	Message string  `json:"message"`
	Target  *Hub    `json:"target"`
	Sender  *Client `json:"sender"`
	Time    string  `json:"time"`
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
