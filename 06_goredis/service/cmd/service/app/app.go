package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomodule/redigo/redis"
	"lectiongoredis/pkg/films"
	"log"
	"net/http"
	"strconv"
	"time"
)

const cacheTimeout = 50 * time.Millisecond

type Server struct {
	mux      chi.Router
	cache    *redis.Pool
	filmsSvc *films.Service
}

func NewServer(mux chi.Router, cache *redis.Pool, filmsSvc *films.Service) *Server {
	return &Server{mux: mux, cache: cache, filmsSvc: filmsSvc}
}

func (s *Server) Init() error {
	s.mux.With(middleware.Logger).Get("/films", s.Top)
	s.mux.With(middleware.Logger).Get("/films/{id}", s.ByID)
	s.mux.With(middleware.Logger).Get("/films/search", s.Search)
	s.mux.With(middleware.Logger).Post("/films", s.Save)

	s.mux.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) Top(writer http.ResponseWriter, request *http.Request) {
	if cached, err := s.FromCache(request.Context(), "films:all"); err == nil {
		log.Print("Got from cache")
		writer.Header().Set("Content-Type", "application/json")
		_, err = writer.Write(cached)
		if err != nil {
			log.Print(err)
		}
		return
	}

	items, err := s.filmsSvc.Top(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(items)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		log.Print(err)
	}

	// После получения данных из основной БД и отправки клиенту, можем сохранить в кэш
	go func() {
		_ = s.ToCache(context.Background(), "films:all", body)
	}()
}

func (s *Server) ByID(writer http.ResponseWriter, request *http.Request) {
	idParam := chi.URLParam(request, "id")
	if idParam == "" {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if cached, err := s.FromCache(request.Context(), fmt.Sprintf("films:%s", idParam)); err == nil {
		log.Print("Got from cache")
		writer.Header().Set("Content-Type", "application/json")
		_, err = writer.Write(cached)
		if err != nil {
			log.Print(err)
		}
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}


	item, err := s.filmsSvc.ByID(request.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(item)
	if errors.Is(err, films.ErrNotFound) {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		log.Print(err)
	}

	// После получения данных из основной БД и отправки клиенту, можем сохранить в кэш
	go func() {
		_ = s.ToCache(context.Background(), fmt.Sprintf("films:%s", idParam), body)
	}()
}

func (s *Server) Search(writer http.ResponseWriter, request *http.Request) {
	rating, err := strconv.ParseFloat(request.URL.Query().Get("min_rating"), 64)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if cached, err := s.FromCache(request.Context(), fmt.Sprintf("films:min_rating:%f", rating)); err == nil {
		log.Print("Got from cache")
		writer.Header().Set("Content-Type", "application/json")
		_, err = writer.Write(cached)
		if err != nil {
			log.Print(err)
		}
		return
	}

	items, err := s.filmsSvc.SearchByRating(request.Context(), rating)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(items)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		log.Print(err)
	}

	// После получения данных из основной БД и отправки клиенту, можем сохранить в кэш
	go func() {
		_ = s.ToCache(context.Background(), fmt.Sprintf("films:min_rating:%f", rating), body)
	}()
}

func (s *Server) Save(writer http.ResponseWriter, request *http.Request) {
	var film films.Film
	err := json.NewDecoder(request.Body).Decode(&film)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	saved, err := s.filmsSvc.Save(request.Context(), &film)
	if errors.Is(err, films.ErrNotFound) {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(saved)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) FromCache(ctx context.Context, key string) ([]byte, error) {
	conn, err := s.cache.GetContext(ctx)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	reply, err := redis.DoWithTimeout(conn, cacheTimeout, "GET", key)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	value, err := redis.Bytes(reply, err)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return value, err
}

func (s *Server) ToCache(ctx context.Context, key string, value []byte) error {
	conn, err := s.cache.GetContext(ctx)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	_, err = redis.DoWithTimeout(conn, cacheTimeout, "SET", key, value)
	if err != nil {
		log.Print(err)
	}
	return err
}
