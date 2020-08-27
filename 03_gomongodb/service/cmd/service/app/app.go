package app

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	mux chi.Router
	db  *mongo.Database
}

type Order struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Start   int                `json:"start"`
	Film    Film               `json:"film"`
	Seats   []Seat             `json:"seats"`
	Price   int64              `json:"price"`
	Created int64              `json:"created"`
}

type Film struct {
	Title    string   `json:"title"`
	Rating   float64  `json:"rating"`
	Cashback float64  `json:"cashback"`
	Genres   []string `json:"genres"`
}

type Seat struct {
	Row    int `json:"row"`
	Number int `json:"number"`
}

func NewServer(mux chi.Router, db *mongo.Database) *Server {
	return &Server{mux: mux, db: db}
}

func (s *Server) Init() error {
	s.mux.With(middleware.Logger).Get("/orders", s.All)
	s.mux.With(middleware.Logger).Get("/orders/{id}", s.ByID)
	s.mux.With(middleware.Logger).Get("/orders/search", s.Search)
	s.mux.With(middleware.Logger).Post("/orders", s.Save)

	s.mux.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	})

	return nil
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) All(writer http.ResponseWriter, request *http.Request) {
	cursor, err := s.db.Collection("orders").Find(request.Context(), bson.D{})
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

	//var orders []Order
	orders := make([]Order, 0)
	for cursor.Next(request.Context()) {
		var order Order
		err = cursor.Decode(&order)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		orders = append(orders, order)
	}
	if err = cursor.Err(); err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(orders)
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

	var order Order
	err = s.db.Collection("orders").FindOne(request.Context(), bson.D{{"_id", id}}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		log.Print(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(order)
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

	cursor, err := s.db.Collection("orders").Find(
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

	orders := make([]Order, 0)
	for cursor.Next(request.Context()) {
		var order Order
		err = cursor.Decode(&order)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		orders = append(orders, order)
	}

	body, err := json.Marshal(orders)
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
	var order Order
	err := json.NewDecoder(request.Body).Decode(&order)
	if err != nil {
		log.Print(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if order.ID == primitive.NilObjectID {
		order.Created = time.Now().Unix()
		result, err := s.db.Collection("orders").InsertOne(
			request.Context(),
			order,
		)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		order.ID = result.InsertedID.(primitive.ObjectID)
	} else {
		result, err := s.db.Collection("orders").UpdateOne(
			request.Context(),
			bson.D{{"_id", order.ID}},
			bson.D{
			{"$set", bson.D{
				{"start", order.Start},
				{"price", order.Price},
			}},
		})
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

	body, err := json.Marshal(order)
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
