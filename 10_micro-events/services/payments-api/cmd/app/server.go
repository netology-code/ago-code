package app

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
	"payments-api/pkg/payments"
	"strconv"
)

type Server struct {
	paymentsSvc *payments.Service
	mux         chi.Router
}

func NewServer(paymentsSvc *payments.Service, mux chi.Router) *Server {
	return &Server{paymentsSvc: paymentsSvc, mux: mux}
}

func (s *Server) Init() error {
	s.mux.Use(middleware.Logger)

	s.mux.Route("/api", func(r chi.Router) {
		r.Post("/payments", s.pay)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) pay(writer http.ResponseWriter, request *http.Request) {
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

	type requestDTO struct {
		Amount    int64  `json:"amount"`
		Category string `json:"category"`
	}

	var reqDTO *requestDTO
	err = json.NewDecoder(request.Body).Decode(&reqDTO)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.paymentsSvc.Pay(request.Context(), userID, reqDTO.Amount, reqDTO.Category)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
