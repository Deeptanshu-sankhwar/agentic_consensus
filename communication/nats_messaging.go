package communication

import (
	"fmt"

	"github.com/Deeptanshu-sankhwar/agentic_consensus/core"
	"github.com/nats-io/nats.go"
)

type Messenger struct {
	broker *core.NATSBroker
}

// Creates a new Messenger instance with the given NATS URL
func NewMessenger(url string) (*Messenger, error) {
	broker, err := core.NewNATSBroker(url)
	if err != nil {
		return nil, err
	}
	return &Messenger{broker: broker}, nil
}

// Publishes a message to a global subject
func (m *Messenger) PublishGlobal(subject, message string) error {
	return m.broker.Publish(subject, []byte(message))
}

// Sends a private message to a specific agent
func (m *Messenger) PublishPrivate(agentID, message string) error {
	subject := fmt.Sprintf("agent.%s.private", agentID)
	return m.broker.Publish(subject, []byte(message))
}

// Subscribes to messages on a global topic
func (m *Messenger) SubscribeGlobal(subject string, handler nats.MsgHandler) error {
	return m.broker.Subscribe(subject, handler)
}

// Subscribes to private messages for a specific agent
func (m *Messenger) SubscribePrivate(agentID string, handler nats.MsgHandler) error {
	subject := fmt.Sprintf("agent.%s.private", agentID)
	return m.broker.Subscribe(subject, handler)
}
