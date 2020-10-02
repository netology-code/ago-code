package main

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"transactions/cmd/app"
	"transactions/pkg/transactions"
)

const (
	defaultPort              = "9999"
	defaultHost              = "0.0.0.0"
	defaultDSN               = "postgres://app:pass@transactionsdb:5432/db"
	defaultBrokerURL         = "kafka:9092"
	defaultTransactionsTopic = "transactions"
	defaultGroup             = "transactions-core"
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

	dsn, ok := os.LookupEnv("APP_DSN")
	if !ok {
		dsn = defaultDSN
	}

	brokerURL, ok := os.LookupEnv("APP_BROKER_URL")
	if !ok {
		brokerURL = defaultBrokerURL
	}

	transactionsTopic, ok := os.LookupEnv("APP_TRANSACTIONS_TOPIC")
	if !ok {
		transactionsTopic = defaultTransactionsTopic
	}

	group, ok := os.LookupEnv("APP_GROUP")
	if !ok {
		group = defaultGroup
	}

	if err := execute(net.JoinHostPort(host, port), dsn, brokerURL, transactionsTopic, group); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string, dsn string, brokerURL string, transactionsTopic string, group string) error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		pool.Close()
	}()

	transactionsSvc := transactions.NewService(pool)
	mux := chi.NewRouter()

	application := app.NewServer(transactionsSvc, mux)
	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}
	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}

	consumerGroup, err := waitForKafka(brokerURL, group)
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

	subscriber := app.NewSubscriber(transactionsSvc, consumerGroup)

	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errs <- err
		}
	}()
	go func() {
		for {
			err := consumerGroup.Consume(context.Background(), []string{transactionsTopic}, subscriber)
			if err != nil {
				errs <- err
				return
			}
		}
	}()

	return <- errs
}

func waitForKafka(brokerURL string, group string) (sarama.ConsumerGroup, error) {
	for {
		select {
		case <-time.After(time.Minute):
			return nil, errors.New("can't connect to kafka")
		default:

		}
		sarama.Logger = log.New(os.Stdout, "", log.Ltime)
		config := sarama.NewConfig()
		config.ClientID = "transactions"
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
		config.Version = sarama.V2_6_0_0

		consumerGroup, err := sarama.NewConsumerGroup([]string{brokerURL}, group, config)
		if err != nil {
			log.Print(err)
			continue
		}

		return consumerGroup, nil
	}
}
