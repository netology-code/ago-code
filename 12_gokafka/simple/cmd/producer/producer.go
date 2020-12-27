package main

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"os"
)

const (
	defaultBrokerURL = "localhost:9093"
	defaultTopic     = "payments"
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

	if err := execute(brokerURL, topic); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

type MessageDTO struct {
	UserID   int64  `json:"userId"`
	Amount   int64  `json:"amount"`
}

func execute(brokerURL string, topic string) error {
	sarama.Logger = log.New(os.Stdout, "kafka", log.Ltime)
	config := sarama.NewConfig()
	config.ClientID = "payments-api"
	config.Producer.Return.Successes = true
	config.Version = sarama.V2_6_0_0

	producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := producer.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	data, err := json.Marshal(MessageDTO{
		UserID:   1,
		Amount:   10_000,
	})
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	message, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Printf("%d %#v", offset, message)

	return nil
}

