package app

import (
	"backend/pkg/auth"
	"backend/pkg/transactions"
	"context"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
)

type Server struct {
	authSvc         *auth.Service
	transactionsSvc *transactions.Service
	mux             chi.Router
}

func NewServer(authSvc *auth.Service, transactionsSvc *transactions.Service, mux chi.Router) *Server {
	return &Server{authSvc: authSvc, transactionsSvc: transactionsSvc, mux: mux}
}

func (s *Server) Init() error {
	s.mux.Use(middleware.Logger)

	s.mux.Route("/api", func(r chi.Router) {
		r.Post("/token", s.token)
		r.With(Auth(func(ctx context.Context, token string) (int64, error) {
			return s.authSvc.Auth(ctx, token)
		})).Get("/transactions", s.transactions)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) token(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		log.Print("can't parse form")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	login := request.PostForm.Get("login")
	if login == "" {
		log.Print("no login in request")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	password := request.PostForm.Get("password")
	if password == "" {
		log.Print("no password in request")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := s.authSvc.Token(request.Context(), login, password)
	if err != nil {
		log.Printf("Auth Service returns error: %v", err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data := &tokenDTO{Token: token}
	respBody, err := json.Marshal(data)
	if err != nil {
		log.Printf("can't marshall data: %v", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(respBody)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) transactions(writer http.ResponseWriter, request *http.Request) {
	userID, err := AuthFrom(request.Context())
	if err != nil {
		log.Printf("can't find userID in context: %v", err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data, err := s.transactionsSvc.Transactions(request.Context(), userID)
	if err != nil {
		log.Printf("Transactions Service returns error: %v", err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
		return
	}
}
