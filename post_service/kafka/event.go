package kafka

import (
	"encoding/json"
	"time"
)

type ClientRegisteredEvent struct {
	ClientID uint32    `json:"client_id"`
	JoinedAt time.Time `json:"joined_at"`
}

func SendClientRegistered(clientID uint32) {
	event := ClientRegisteredEvent{
		ClientID: clientID,
		JoinedAt: time.Now(),
	}
	payload, _ := json.Marshal(event)
	SendMessage("client_registered", "client", payload)
}

type InteractionEvent struct {
	ClientID uint32    `json:"client_id"`
	EntityID uint32    `json:"entity_id"`
	ActionAt time.Time `json:"action_at"`
}

func SendInteractionEvent(topic string, clientID, entityID uint32) {
	event := InteractionEvent{
		ClientID: clientID,
		EntityID: entityID,
		ActionAt: time.Now(),
	}
	payload, _ := json.Marshal(event)
	SendMessage(topic, "interaction", payload)
}
