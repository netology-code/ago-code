package app

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
)

type Server struct {
	mux    chi.Router
}

func NewServer(mux chi.Router) *Server {
	return &Server{mux: mux}
}

func (s *Server) Init() error {
	s.mux.Use(middleware.Logger)

	s.mux.Route("/api", func(r chi.Router) {
		r.Get("/transactions", s.transactions)
		r.Get("/health", s.health)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) transactions(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	_, err := writer.Write([]byte(`
[
	{
		"id": 1,
		"amount": "1000"
	},
	{
		"id": 2,
		"amount": "2000"
	}
]
	`))
	if err != nil {
		log.Print(err)
		return
	}
}

func (s *Server) health(writer http.ResponseWriter, request *http.Request) {
	log.Print("status OK")
	writer.WriteHeader(http.StatusOK)
}
