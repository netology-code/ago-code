package app

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
	"strconv"
	"transactions/pkg/transactions"
)

type Server struct {
	transactionsSvc *transactions.Service
	mux             chi.Router
}

func NewServer(transactionsSvc *transactions.Service, mux chi.Router) *Server {
	return &Server{transactionsSvc: transactionsSvc, mux: mux}
}

func (s *Server) Init() error {
	s.mux.Use(middleware.Logger)

	s.mux.Route("/api", func(r chi.Router) {
		r.Get("/transactions", s.transactions)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) transactions(writer http.ResponseWriter, request *http.Request) {
	userIDHeader := request.Header.Get("X-UserID")
	if userIDHeader == "" {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDHeader, 10, 64)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	type Transaction struct {
		ID       int64  `json:"id"`
		UserID   int64  `json:"userId"`
		Category string `json:"category"`
		Amount   int64  `json:"amount"`
		Created  int64  `json:"created"`
	}
	type responseDTO []Transaction
	var respDTO responseDTO

	records, err := s.transactionsSvc.Transactions(request.Context(), userID)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for _, record := range records {
		respDTO = append(respDTO, Transaction{
			ID:       record.ID,
			UserID:   record.UserID,
			Category: record.Category,
			Amount:   record.Amount,
			Created:  record.Created,
		})
	}

	data, err := json.Marshal(respDTO)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}
