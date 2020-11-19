package authorizator

import (
	"context"
	"net/http"
)


type HasAnyRoleFunc func(ctx context.Context, roles ...string) bool

func Authorizator(hasAnyRole HasAnyRoleFunc, roles ...string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !hasAnyRole(request.Context(), roles...) {
				writer.WriteHeader(http.StatusForbidden)
				return
			}

			handler.ServeHTTP(writer, request)
		})
	}
}
