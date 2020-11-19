package main

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"lectiongoredis/cmd/service/app"
	"lectiongoredis/pkg/films"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultPort     = "9999"
	defaultHost     = "0.0.0.0"
	defaultDbDSN    = "postgres://app:pass@localhost:5432/db"
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

	dbDSN, ok := os.LookupEnv("APP_DSN")
	if !ok {
		dbDSN = defaultDbDSN
	}

	cacheDSN, ok := os.LookupEnv("APP_CACHE_DSN")
	if !ok {
		cacheDSN = defaultCacheDSN
	}

	if err := execute(net.JoinHostPort(host, port), dbDSN, cacheDSN); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string, dbDSN string, cacheDSN string) error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		return err
	}
	defer pool.Close()

	filmsSvc := films.NewService(pool)
	mux := chi.NewRouter()

	cache := &redis.Pool{
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.DialURL(cacheDSN)
		},
	}
	defer func() {
		if cerr := cache.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	application := app.NewServer(mux, cache, filmsSvc)
	err = application.Init()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}
