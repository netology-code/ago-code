package main

import (
	"context"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"os"
	"service/cmd/app"
)

const (
	defaultPort = "9999"
	defaultHost = "0.0.0.0"
	defaultDSN  = "mongodb://app:pass@localhost:27017/" + defaultDB
	defaultDB   = "db"
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

	db, ok := os.LookupEnv("APP_DB")
	if !ok {
		db = defaultDB
	}

	if err := execute(net.JoinHostPort(host, port), dsn, db); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, dsn string, db string) error {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		log.Print(err)
		return err
	}

	database := client.Database(db)

	mux := chi.NewMux()

	application := app.NewServer(mux, database)
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
