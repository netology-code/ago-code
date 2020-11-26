package main

import (
	"backend/cmd/app"
	"backend/pkg/auth"
	"backend/pkg/transactions"
	"github.com/go-chi/chi"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
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

	tokenLifeTime := time.Hour

	publicKey, err := ioutil.ReadFile("keys/public.key")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	transactionsAPIURL, ok := os.LookupEnv("APP_TRANSACTIONS_URL")
	if !ok {
		transactionsAPIURL = defaultTransactionsAPIURL
	}

	if err := execute(net.JoinHostPort(host, port), publicKey, tokenLifeTime, transactionsAPIURL); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, publicKey []byte, tokenLifeTime time.Duration, transactionsAPIURL string) error {
	authSvc := auth.NewService(publicKey, tokenLifeTime)
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


