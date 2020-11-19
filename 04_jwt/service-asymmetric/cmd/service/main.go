package main

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4/pgxpool"
	"io/ioutil"
	"log"
	"net/http"
	"service/cmd/service/app"
	"time"

	//"context"
	//"github.com/jackc/pgx/v4/pgxpool"
	//"log"
	//"net"
	//"net/http"
	//"os"
	//"service/cmd/service/app"
	//"service/pkg/business"
	//"service/pkg/security"
	"net"
	"os"
	"service/pkg/business"
	"service/pkg/security"
)

const (
	defaultPort = "9999"
	defaultHost = "0.0.0.0"
	defaultDSN  = "postgres://app:pass@localhost:5432/db"
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

	tokenLifeTime := time.Hour

	privateKey, err := ioutil.ReadFile("keys/private.key")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	publicKey, err := ioutil.ReadFile("keys/public.key")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	if err := execute(net.JoinHostPort(host, port), dsn, privateKey, publicKey, tokenLifeTime); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string, dsn string, privateKey, publicKey []byte, tokenLifeTime time.Duration) error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Print(err)
		return err
	}
	defer pool.Close()

	securitySvc := security.NewService(pool, privateKey, publicKey, tokenLifeTime)
	businessSvc := business.NewService(pool)
	router := chi.NewRouter()
	application := app.NewServer(securitySvc, businessSvc, router)
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
