package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"lectiongoredis/cmd/service/app/middleware/cache"
	"log"
	"net/http"
	"strconv"
	"time"
)

const cacheTimeout = 50 * time.Millisecond

type Server struct {
	mux   chi.Router
	db    *mongo.Database
	cache *redis.Pool
}

type Film struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title    string             `json:"title"`
	Rating   float64            `json:"rating"`
	Cashback float64            `json:"cashback"`
	Genres   []string           `json:"genres"`
	Start    int64              `json:"start"`
}

func NewServer(mux chi.Router, db *mongo.Database, cache *redis.Pool) *Server {
	return &Server{mux: mux, db: db, cache: cache}
}

func (s *Server) Init() error {
	cacheMd := cache.Cache(func(ctx context.Context, path string) ([]byte, error) {
		value, err := s.FromCache(ctx, path)
		if err != nil && errors.Is(err, redis.ErrNil) {
			return nil, cache.ErrNotInCache
		}
		return value, err
	}, func(ctx context.Context, path string, data []byte) error {
		return s.ToCache(context.Background(), path, data)
	})

	s.mux.With(middleware.Logger, cacheMd).Get("/cached/films", s.All)
	s.mux.With(middleware.Logger, cacheMd).Get("/cached/films/{id}", s.ByID)
	s.mux.With(middleware.Logger).Get("/films", s.All)
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

func (s *Server) All(writer http.ResponseWriter, request *http.Request) {
	cursor, err := s.db.Collection("films").Find(request.Context(), bson.D{})
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		if cerr := cursor.Close(request.Context()); cerr != nil {
			log.Print(cerr)
		}
	}()

	films := make([]Film, 0)
	for cursor.Next(request.Context()) {
		var film Film
		err = cursor.Decode(&film)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		films = append(films, film)
	}
	if err = cursor.Err(); err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(films)
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

func (s *Server) ByID(writer http.ResponseWriter, request *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(request, "id"))
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var film Film
	err = s.db.Collection("films").FindOne(request.Context(), bson.D{{"_id", id}}).Decode(&film)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(film)
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

func (s *Server) Search(writer http.ResponseWriter, request *http.Request) {
	rating, err := strconv.ParseFloat(request.URL.Query().Get("min_rating"), 64)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	cursor, err := s.db.Collection("films").Find(
		request.Context(),
		bson.D{
			{"film.rating", bson.D{
				{"$gt", rating},
			}},
		},
	)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		if cerr := cursor.Close(request.Context()); cerr != nil {
			log.Print(cerr)
		}
	}()

	films := make([]Film, 0)
	for cursor.Next(request.Context()) {
		var film Film
		err = cursor.Decode(&film)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		films = append(films, film)
	}

	body, err := json.Marshal(films)
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

func (s Server) Save(writer http.ResponseWriter, request *http.Request) {
	var film Film
	err := json.NewDecoder(request.Body).Decode(&film)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if film.ID == primitive.NilObjectID {
		result, err := s.db.Collection("films").InsertOne(
			request.Context(),
			film,
		)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		film.ID = result.InsertedID.(primitive.ObjectID)
	} else {
		result, err := s.db.Collection("films").ReplaceOne(request.Context(), bson.D{{"_id", film.ID}}, film)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if result.MatchedCount == 0 {
			log.Print(err)
			writer.WriteHeader(http.StatusNotFound)
			return
		}
	}

	body, err := json.Marshal(film)
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
