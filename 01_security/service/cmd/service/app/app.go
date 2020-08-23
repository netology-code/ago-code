package app

import (
	"context"
	"encoding/json"
	"github.com/netology-code/remux/pkg/middleware/authenticator"
	"github.com/netology-code/remux/pkg/middleware/authorizator"
	"github.com/netology-code/remux/pkg/middleware/logger"
	"github.com/netology-code/remux/pkg/remux"
	"log"
	"net/http"
	"service/cmd/service/app/dto"
	"service/cmd/service/app/middleware/identificator"
	"service/pkg/business"
	"service/pkg/security"
)

type Server struct {
	securitySvc *security.Service
	businessSvc *business.Service
	mux         *remux.ReMux
}

func NewServer(securitySvc *security.Service, businessSvc *business.Service, mux *remux.ReMux) *Server {
	return &Server{securitySvc: securitySvc, businessSvc: businessSvc, mux: mux}
}

func (s *Server) Init() error {
	logMd := logger.Logger
	identificatorMd := identificator.Identificator
	authenticatorMd := authenticator.Authenticator(identificator.Identifier, s.securitySvc.UserDetails)

	// функция-связка между middleware и security service (для чистоты security service, который ничего не знает об http)
	roleChecker := func(ctx context.Context, roles ...string) bool {
		userDetails, err := authenticator.Authentication(ctx)
		if err != nil {
			return false
		}
		return s.securitySvc.HasAnyRole(ctx, userDetails, roles...)
	}
	adminRoleMd := authorizator.Authorizator(roleChecker, security.RoleAdmin)
	userRoleMd := authorizator.Authorizator(roleChecker, security.RoleUser)

	if err := s.mux.RegisterPlain(remux.POST, "/login", http.HandlerFunc(s.login), logMd); err != nil {
		return err
	}
	if err := s.mux.RegisterPlain(remux.GET, "/public", http.HandlerFunc(s.public), logMd); err != nil {
		return err
	}
	if err := s.mux.RegisterPlain(remux.GET, "/admin", http.HandlerFunc(s.admin), adminRoleMd, authenticatorMd, identificatorMd, logMd); err != nil {
		return err
	}
	if err := s.mux.RegisterPlain(remux.GET, "/user", http.HandlerFunc(s.user), userRoleMd, authenticatorMd, identificatorMd, logMd); err != nil {
		return err
	}

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) login(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	login := request.PostForm.Get("login")
	if login == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	password := request.PostForm.Get("password")
	if password == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := s.securitySvc.Login(request.Context(), login, password)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data := &dto.TokenDTO{Token: token}
	respBody, err := json.Marshal(data)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(respBody)
	if err != nil {
		log.Print(err)
	}
}

// Доступно всем
func (s *Server) public(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("public"))
}

// Только пользователям с ролью ADMIN
func (s *Server) admin(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("admin"))
}

// Только пользователям с ролью USER
func (s *Server) user(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("user"))
}
