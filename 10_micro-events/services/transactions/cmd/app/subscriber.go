package app

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"transactions/pkg/transactions"
)

type Subscriber struct {
	transactionsSvc *transactions.Service
	consumerGroup   sarama.ConsumerGroup
}

func NewSubscriber(transactionsSvc *transactions.Service, consumerGroup sarama.ConsumerGroup) *Subscriber {
	return &Subscriber{transactionsSvc: transactionsSvc, consumerGroup: consumerGroup}
}

func (s *Subscriber) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (s *Subscriber) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (s *Subscriber) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	type transactionDTO struct {
		ID       string `json:"id"`
		UserID   int64  `json:"userId"`
		Category string `json:"category"`
		Amount   int64  `json:"amount"`
		Created  int64  `json:"created"`
	}

	for message := range claim.Messages() {
		// считаем что тут отправили соо
		log.Printf("received message from topic: %s %s", claim.Topic(), message.Value)
		session.MarkMessage(message, "")

		var transaction *transactionDTO
		err := json.Unmarshal(message.Value, &transaction)
		if err != nil {
			log.Print(err)
			continue
		}

		log.Printf("save message to db: %s", transaction.ID)
		err = s.transactionsSvc.Register(
			context.Background(),
			transactions.Transaction{
				ID:       transaction.ID,
				UserID:   transaction.UserID,
				Category: transaction.Category,
				Amount:   transaction.Amount,
				Created:  transaction.Created,
			})
		if err != nil {
			log.Print(err)
			continue
		}
	}
	return nil
}
