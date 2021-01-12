package payments

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
)

type Service struct {
	producer sarama.SyncProducer
	topic    string
}

func NewService(producer sarama.SyncProducer, topic string) *Service {
	return &Service{producer: producer, topic: topic}
}

type MessageDTO struct {
	UserID   int64  `json:"userId"`
	Amount   int64  `json:"amount"`
	Category string `json:"category"`
}

func (s *Service) Pay(ctx context.Context, userID int64, amount int64, category string) error {
	data, err := json.Marshal(MessageDTO{
		UserID:   userID,
		Amount:   amount,
		Category: category,
	})
	if err != nil {
		return err
	}

	log.Printf("send message to topic: %s", s.topic)
	msg := &sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = s.producer.SendMessage(msg)
	return err
}
