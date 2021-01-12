package app

import (
	"context"
	"errors"
	"net/http"
)

var ErrNoAuth = errors.New("no auth in context")

type AuthFunc func(ctx context.Context, token string) (userID int64, err error)

var authContextKey = &contextKey{"auth context"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

func Auth(authFunc AuthFunc) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			token := request.Header.Get("Authorization")
			if token == "" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			auth, err := authFunc(request.Context(), token)
			if err != nil {
				// упрощённый вариант, нужно ещё добавить проверку на то, что удалённый сервис "отвалился"
				writer.WriteHeader(http.StatusForbidden)
				return
			}

			ctx := context.WithValue(request.Context(), authContextKey, auth)
			request = request.WithContext(ctx)
			handler.ServeHTTP(writer, request)
		})
	}
}

func AuthFrom(ctx context.Context) (int64, error) {
	if value := ctx.Value(authContextKey); value != nil {
		if id, ok := value.(int64); ok {
			return id, nil
		}
	}
	return 0, ErrNoAuth
}
