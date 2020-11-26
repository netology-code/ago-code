package app

import (
	"auth/pkg/auth"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
)

type Server struct {
	authSvc *auth.Service
	mux     chi.Router
}

func NewServer(authSvc *auth.Service, mux chi.Router) *Server {
	return &Server{authSvc: authSvc, mux: mux}
}

func (s *Server) Init() error {
	s.mux.Use(middleware.RealIP)

	s.mux.Route("/api", func(r chi.Router) {
		r.Post("/token", s.token)
		r.Post("/id", s.id)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) token(writer http.ResponseWriter, request *http.Request) {
	// for simplicity just define locally
	type requestDTO struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	type responseDTO struct {
		Token string `json:"token"`
	}

	var reqDTO requestDTO
	err := json.NewDecoder(request.Body).Decode(&reqDTO)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := s.authSvc.Login(request.Context(), reqDTO.Login, reqDTO.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrInvalidPass) {
			http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respDTO := responseDTO{Token: token}
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

// Доступно всем
func (s *Server) id(writer http.ResponseWriter, request *http.Request) {
	// for simplicity just define locally
	type requestDTO struct {
		Token string `json:"token"`
	}

	type responseDTO struct {
		UserID int64 `json:"userId"`
	}

	var reqDTO requestDTO
	err := json.NewDecoder(request.Body).Decode(&reqDTO)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := s.authSvc.UserID(request.Context(), reqDTO.Token)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrInvalidPass) {
			http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respDTO := responseDTO{UserID: userID}
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
