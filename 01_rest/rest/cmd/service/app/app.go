package app

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"rest/pkg/offers"
	"strconv"
)

type Server struct {
	offersSvc *offers.Service
	router    chi.Router
}

func NewServer(offersSvc *offers.Service, router chi.Router) *Server {
	return &Server{offersSvc: offersSvc, router: router}
}

func (s *Server) Init() error {
	s.router.Get("/offers", s.handleGetOffers)
	s.router.Get("/offers/{id}", s.handleGetOfferByID)
	s.router.Post("/offers", s.handleSaveOffer)
	s.router.Delete("/offers/{id}", s.handleRemoveOfferByID)

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) handleGetOffers(writer http.ResponseWriter, request *http.Request) {
	items, err := s.offersSvc.All(request.Context())
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(items)
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

func (s *Server) handleGetOfferByID(writer http.ResponseWriter, request *http.Request) {
	idParam := chi.URLParam(request, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.offersSvc.ByID(request.Context(), id)
	if err != nil {
		// TODO: handle not found
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
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

func (s *Server) handleSaveOffer(writer http.ResponseWriter, request *http.Request) {
	itemToSave := &offers.Offer{}
	err := json.NewDecoder(request.Body).Decode(&itemToSave)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.offersSvc.Save(request.Context(), itemToSave)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
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

func (s *Server) handleRemoveOfferByID(writer http.ResponseWriter, request *http.Request) {
	panic("not implemented")
}
