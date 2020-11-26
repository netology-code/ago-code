package main

import (
	"github.com/go-chi/chi"
	"log"
	"net"
	"net/http"
	"os"
	"transactions-api/cmd/app"
	"transactions-api/pkg/transactions"
)

const (
	defaultPort = "9999"
	defaultHost = "0.0.0.0"
	defaultTransactionsURL = "http://transactions:9999"
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

	transactionsURL, ok := os.LookupEnv("APP_TRANSACTIONS_URL")
	if !ok {
		transactionsURL = defaultTransactionsURL
	}

	if err := execute(net.JoinHostPort(host, port), transactionsURL); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, transactionsURL string) error {
	transactionsSvc := transactions.NewService(&http.Client{}, transactionsURL)

	mux := chi.NewRouter()

	application := app.NewServer(transactionsSvc, mux)
	err := application.Init()
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
