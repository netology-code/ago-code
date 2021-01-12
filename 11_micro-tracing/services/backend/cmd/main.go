package main

import (
	"backend/cmd/app"
	"backend/pkg/auth"
	"backend/pkg/transactions"
	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/go-chi/chi"
	"github.com/hashicorp/consul/api"
	"go.opencensus.io/trace"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
)

const (
	defaultPort      = "9999"
	defaultHost      = "backend"
	defaultConsulURL = "http://consul:8500"
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

	consulURL, ok := os.LookupEnv("APP_CONSUL_URL")
	if !ok {
		consulURL = defaultConsulURL
	}

	if err := execute(host, port, consulURL); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(host string, port string, consulURL string) error {
	parsed, err := url.Parse(consulURL)
	if err != nil {
		return err
	}

	client, err := api.NewClient(&api.Config{
		Address: parsed.Host,
		Scheme:  parsed.Scheme,
	})
	if err != nil {
		return err
	}

	authSvc := auth.NewService(&http.Client{})
	transactionsSvc := transactions.NewService(&http.Client{})

	mux := chi.NewRouter()

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	exporter, _ := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     "jaeger:6831",
		CollectorEndpoint: "http://jaeger:14268/api/traces",
	})
	trace.RegisterExporter(exporter)

	application := app.NewServer(authSvc, transactionsSvc, mux)
	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}

	server := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: application,
	}

	watcher := app.NewWatcher(client)

	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errs <- err
		}
	}()
	go func() {
		err := watcher.Watch([]string{"auth", "transactions"})
		if err != nil {
			errs <- err
		}
	}()

	return <-errs
}
