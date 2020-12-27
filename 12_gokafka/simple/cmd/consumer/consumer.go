package main

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"os"
)

const (
	defaultBrokerURL = "localhost:9093"
	defaultTopic     = "payments"
	defaultGroup     = "payments-gateway"
)

func main() {
	brokerURL, ok := os.LookupEnv("APP_BROKER_URL")
	if !ok {
		brokerURL = defaultBrokerURL
	}

	topic, ok := os.LookupEnv("APP_TOPIC")
	if !ok {
		topic = defaultTopic
	}

	group, ok := os.LookupEnv("APP_GROUP")
	if !ok {
		group = defaultGroup
	}

	if err := execute(brokerURL, topic, group); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(brokerURL string, topic string, group string) error {
	sarama.Logger = log.New(os.Stdout, "kafka", log.Ltime)
	config := sarama.NewConfig()
	config.ClientID = "payments-gateway"
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // -2
	config.Consumer.Offsets.AutoCommit.Enable = false     // отключаем auto commit об обработке
	config.Version = sarama.V2_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup([]string{brokerURL}, group, config)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := consumerGroup.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	handler := &ConsumerHandler{}
	// бесконечный цикл, если будет перенастроен топик, то будем переподписываться
	for {
		err = consumerGroup.Consume(context.Background(), []string{topic}, handler)
		if err != nil {
			return err
		}
	}
}

type MessageDTO struct {
	UserID int64 `json:"userId"`
	Amount int64 `json:"amount"`
}

type ConsumerHandler struct { }

func (h *ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim,
) error {
	// цикл обработки сообщений, канал будет закрыт при серверной ребалансировке
	for message := range claim.Messages() {
		log.Printf("received message from topic: %s, %s", claim.Topic(), message.Value)

		var dto *MessageDTO
		err := json.Unmarshal(message.Value, &dto)
		if err != nil {
			log.Print(err)
			continue
		}

		session.MarkMessage(message, "")
		session.Commit()
	}

	return nil
}

