package logger

import (
	"log"
	"net/http"
)

func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("new request: %s %s", request.Method, request.URL.Path)
		handler.ServeHTTP(writer, request)
	})
}
