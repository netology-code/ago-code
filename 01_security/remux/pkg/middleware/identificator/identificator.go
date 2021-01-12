package identificator

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var ErrNoIdentifier = errors.New("no identifier")

var identifierContextKey = &contextKey{"identifier context"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

func Identificator(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// naive realization
		parts := strings.Split(request.RemoteAddr, ":")
		if len(parts) == 2 {
			ctx := context.WithValue(request.Context(), identifierContextKey, &parts[0])
			request = request.WithContext(ctx)
		}

		handler.ServeHTTP(writer, request)
	})
}

func Identifier(ctx context.Context) (*string, error) {
	value, ok := ctx.Value(identifierContextKey).(*string)
	if !ok {
		return nil, ErrNoIdentifier
	}
	return value, nil
}

