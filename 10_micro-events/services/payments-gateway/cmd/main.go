package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"log"
	"os"
	"time"
)

const (
	defaultBrokerURL         = "kafka:9092"
	defaultPaymentsTopic     = "payments"
	defaultTransactionsTopic = "transactions"
	defaultGroup             = "payments-gateway"
)

func main() {
	brokerURL, ok := os.LookupEnv("APP_BROKER_URL")
	if !ok {
		brokerURL = defaultBrokerURL
	}

	paymentsTopic, ok := os.LookupEnv("APP_PAYMENTS_TOPIC")
	if !ok {
		paymentsTopic = defaultPaymentsTopic
	}

	transactionsTopic, ok := os.LookupEnv("APP_TRANSACTIONS_TOPIC")
	if !ok {
		transactionsTopic = defaultTransactionsTopic
	}

	group, ok := os.LookupEnv("APP_GROUP")
	if !ok {
		group = defaultGroup
	}

	if err := execute(brokerURL, paymentsTopic, transactionsTopic, group); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(brokerURL string, paymentTopic string, transactionsTopic string, group string) error {
	// только один из группы будет получать сообщения
	consumerGroup, producer, err := waitForKafka(brokerURL, group)
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

	defer func() {
		if cerr := producer.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	handler := &Handler{Producer: producer, TransactionsTopic: transactionsTopic}
	for {
		err = consumerGroup.Consume(context.Background(), []string{paymentTopic}, handler)
		if err != nil {
			return err
		}
	}
}

func waitForKafka(brokerURL string, group string) (sarama.ConsumerGroup, sarama.SyncProducer, error) {
	for {
		select {
		case <-time.After(time.Minute):
			return nil, nil, errors.New("can't connect to kafka")
		default:

		}
		sarama.Logger = log.New(os.Stdout, "", log.Ltime)
		config := sarama.NewConfig()
		config.ClientID = "payments-gateway"
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
		config.Producer.Return.Successes = true
		config.Version = sarama.V2_6_0_0

		consumerGroup, err := sarama.NewConsumerGroup([]string{brokerURL}, group, config)
		if err != nil {
			log.Print(err)
			continue
		}

		producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
		if err != nil {
			log.Print(err)
			continue
		}

		return consumerGroup, producer, nil
	}
}

// обрабатывает event и кладёт обратно новый event
type Handler struct {
	Producer          sarama.SyncProducer
	TransactionsTopic string
	Ready             chan struct{}
}

func (h *Handler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	type paymentDTO struct {
		UserID   int64  `json:"userId"`
		Amount   int64  `json:"amount"`
		Category string `json:"category"`
	}

	type transactionDTO struct {
		ID       string `json:"id"`
		UserID   int64  `json:"userId"`
		Category string `json:"category"`
		Amount   int64  `json:"amount"`
		Created  int64  `json:"created"`
	}

	for message := range claim.Messages() {
		// считаем что тут отправили сообщение в платёжную систему
		log.Printf("received message from topic: %s, %s", claim.Topic(), message.Value)
		session.MarkMessage(message, "")

		var payment *paymentDTO
		err := json.Unmarshal(message.Value, &payment)
		if err != nil {
			log.Print(err)
			continue
		}

		transaction := &transactionDTO{
			ID:       uuid.New().String(),
			UserID:   payment.UserID,
			Category: payment.Category,
			Amount:   payment.Amount,
			Created:  time.Now().Unix(),
		}

		data, err := json.Marshal(transaction)
		if err != nil {
			log.Print(err)
			continue
		}

		log.Printf("sent message to topic: %s", h.TransactionsTopic)
		_, _, err = h.Producer.SendMessage(&sarama.ProducerMessage{
			Topic: h.TransactionsTopic,
			Value: sarama.ByteEncoder(data),
		})
		if err != nil {
			log.Print(err)
		}
	}

	return nil
}
