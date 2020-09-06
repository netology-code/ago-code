package main

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"lectiongoredis/cmd/service/app"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultPort     = "9999"
	defaultHost     = "0.0.0.0"
	defaultDSN      = "mongodb://app:pass@localhost:27017/" + defaultDB
	defaultDB       = "db"
	defaultCacheDSN = "redis://localhost:6379/0"
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

	cacheDSN, ok := os.LookupEnv("APP_CACHE_DSN")
	if !ok {
		cacheDSN = defaultCacheDSN
	}

	if err := execute(net.JoinHostPort(host, port), dsn, db, cacheDSN); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, dsn string, db string, cacheDSN string) error {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		log.Print(err)
		return err
	}

	database := client.Database(db)
	
	mux := chi.NewMux()

	cache := &redis.Pool{
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.DialURL(cacheDSN)
		},
	}

	application := app.NewServer(mux, database, cache)
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