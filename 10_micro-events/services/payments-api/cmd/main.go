package main

import (
	"errors"
	"github.com/Shopify/sarama"
	"github.com/go-chi/chi"
	"log"
	"net"
	"net/http"
	"os"
	"payments-api/cmd/app"
	"payments-api/pkg/payments"
	"time"
)

const (
	defaultPort = "9999"
	defaultHost = "0.0.0.0"
	defaultBrokerURL = "kafka:9092"
	defaultTopic = "payments"
)

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	brokerURL, ok := os.LookupEnv("APP_BROKER_URL")
	if !ok {
		brokerURL = defaultBrokerURL
	}

	topic, ok := os.LookupEnv("APP_TOPIC")
	if !ok {
		topic = defaultTopic
	}

	if err := execute(net.JoinHostPort(host, port), brokerURL, topic); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string, brokerURL string, topic string) error {
	producer, err := waitForKafka(brokerURL)
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

	paymentsSvc := payments.NewService(producer, topic)

	mux := chi.NewRouter()

	application := app.NewServer(paymentsSvc, mux)
	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}

func waitForKafka(brokerURL string) (sarama.SyncProducer, error) {
	for {
		select {
		case <- time.After(time.Minute):
			return nil, errors.New("can't connect to kafka")
		default:

		}
		sarama.Logger = log.New(os.Stdout, "", log.Ltime)
		config := sarama.NewConfig()
		config.ClientID = "payments-api"
		config.Producer.Return.Successes = true
		config.Version = sarama.V2_6_0_0

		producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
		if err != nil {
			log.Print(err)
			time.Sleep(time.Second)
			continue
		}

		return producer, nil
	}
}