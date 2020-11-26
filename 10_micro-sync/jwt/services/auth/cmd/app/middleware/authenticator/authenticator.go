package authenticator

import (
	"context"
	"errors"
	"net/http"
)

var ErrNoAuthentication = errors.New("no authentication")

var authenticationContextKey = &contextKey{"authentication context"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

type IdentifierFunc func(ctx context.Context) (*string, error)

type UserDetailsFunc func(ctx context.Context, id *string) (interface{}, error)

func Authenticator(identifier IdentifierFunc, userDetails UserDetailsFunc) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			id, err := identifier(request.Context())
			if err != nil {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			userDetails, err := userDetails(request.Context(), id)
			if err != nil {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(request.Context(), authenticationContextKey, userDetails)
			request = request.WithContext(ctx)

			handler.ServeHTTP(writer, request)
		})
	}
}

func Authentication(ctx context.Context) (interface{}, error) {
	if value := ctx.Value(authenticationContextKey); value != nil {
		return value, nil
	}
	return nil, ErrNoAuthentication
}

