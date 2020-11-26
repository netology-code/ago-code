package app

import (
	"backend/cmd/app/middleware/authenticator"
	"backend/cmd/app/middleware/authorizator"
	"backend/cmd/app/middleware/identificator"
	"backend/pkg/auth"
	"backend/pkg/transactions"
	"context"
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

	identificatorMd := identificator.Identificator
	authenticatorMd := authenticator.Authenticator(
		identificator.Identifier, s.authSvc.UserDetails,
	)

	// функция-связка между middleware и security service
	// (для чистоты security service ничего не знает об http)
	roleChecker := func(ctx context.Context, roles ...string) bool {
		userDetails, err := authenticator.Authentication(ctx)
		if err != nil {
			return false
		}
		return s.authSvc.HasAnyRole(ctx, userDetails, roles...)
	}
	userRoleMd := authorizator.Authorizator(roleChecker, auth.RoleUser)

	s.mux.Route("/api", func(r chi.Router) {
		r.With(identificatorMd, authenticatorMd, userRoleMd).Get("/transactions", s.transactions)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) transactions(writer http.ResponseWriter, request *http.Request) {
	profile, err := authenticator.Authentication(request.Context())
	if err != nil {
		log.Printf("can't find userID in context: %v", err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	details, ok := profile.(*auth.UserDetails)
	if !ok {
		log.Printf("can't get details from context: %v", details)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// for simplicity send just userID, but we can send whole token or details
	data, err := s.transactionsSvc.Transactions(request.Context(), details.UserID)
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
