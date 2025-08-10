package services

import (
	"log"

	"github.com/nats-io/nats.go"
)

var (
	NC             *nats.Conn
	PaymentSubject = "menageria"
)

func InitNATS(natsURL string) {
	var err error
	NC, err = nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to NATS server")
}

func PublishMessage(subject string, message []byte) {
	if err := NC.Publish(subject, message); err != nil {
		log.Fatalf("Error publishing message: %v", err)
	}
}

func SubscribeToSubject(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	subscription, err := NC.Subscribe(subject, handler)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}
