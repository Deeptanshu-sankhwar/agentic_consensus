package core

import (
	"log"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var NatsBrokerInstance *nats.Conn
var natsServer *server.Server

// SetupNATS establishes a connection to NATS server or starts an embedded one if connection fails
func SetupNATS(natsURL string) {
	var err error
	NatsBrokerInstance, err = nats.Connect(natsURL)
	if err != nil {
		log.Printf("Could not connect to NATS at %s, starting embedded server...", natsURL)
		opts := &server.Options{
			Port:   4222,
			Host:   "localhost",
			NoLog:  false,
			NoSigs: true,
		}

		natsServer, _ = server.NewServer(opts)
		go natsServer.Start()

		if !natsServer.ReadyForConnections(4 * time.Second) {
			log.Fatal("NATS server failed to start")
		}
		log.Println("Started embedded NATS server on port 4222")

		NatsBrokerInstance, err = nats.Connect("nats://localhost:4222")
		if err != nil {
			log.Fatalf("Failed to connect to embedded NATS: %v", err)
		}
	}
	log.Printf("Connected to NATS at %s", natsURL)
}

// CloseNATS closes the NATS connection and shuts down the server if it was embedded
func CloseNATS() {
	if NatsBrokerInstance != nil {
		NatsBrokerInstance.Close()
	}
	if natsServer != nil {
		natsServer.Shutdown()
	}
}

type NATSBroker struct {
	Conn *nats.Conn
}

// NewNATSBroker creates a new NATS broker instance with the specified URL
func NewNATSBroker(url string) (*NATSBroker, error) {
	nc, err := nats.Connect(url,
		nats.Timeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return &NATSBroker{Conn: nc}, nil
}

// Publish sends data to the specified NATS subject
func (b *NATSBroker) Publish(subject string, data []byte) error {
	log.Printf("Sending data to %s", subject)
	return b.Conn.Publish(subject, data)
}

// Subscribe registers a message handler for the specified subject
func (b *NATSBroker) Subscribe(subject string, cb nats.MsgHandler) error {
	_, err := b.Conn.Subscribe(subject, cb)
	return err
}

// Close terminates the NATS connection
func (b *NATSBroker) Close() {
	b.Conn.Close()
}
