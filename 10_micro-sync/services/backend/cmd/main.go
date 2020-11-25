package main

import (
	"backend/cmd/app"
	"backend/pkg/auth"
	"backend/pkg/transactions"
	"github.com/go-chi/chi"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultPort               = "9999"
	defaultHost               = "0.0.0.0"
	defaultAuthURL            = "http://auth:9999"
	defaultTransactionsAPIURL = "http://transactions-api:9999"
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

	authURL, ok := os.LookupEnv("APP_AUTH_URL")
	if !ok {
		authURL = defaultAuthURL
	}

	transactionsAPIURL, ok := os.LookupEnv("APP_TRANSACTIONS_URL")
	if !ok {
		transactionsAPIURL = defaultTransactionsAPIURL
	}

	if err := execute(net.JoinHostPort(host, port), authURL, transactionsAPIURL); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, authURL string, transactionsAPIURL string) error {
	authSvc := auth.NewService(&http.Client{}, authURL)
	transactionsSvc := transactions.NewService(&http.Client{}, transactionsAPIURL)

	mux := chi.NewRouter()

	application := app.NewServer(authSvc, transactionsSvc, mux)
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


