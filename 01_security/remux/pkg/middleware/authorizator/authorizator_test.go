package authorizator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"github.com/netology-code/remux/pkg/remux"
	"testing"
)

func TestAuthorizator(t *testing.T) {
	mux := remux.NewReMux()
	roleAdminMd := Authorizator(func(ctx context.Context, roles ...string) bool {
		return false
	}, "ADMIN")
	if err := mux.RegisterPlain(
		remux.GET,
		"/get",
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) { }),
		roleAdminMd,
	); err != nil {
		t.Fatal(err)
	}

	type args struct {
		method remux.Method
		path   string
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "GET", args: args{method: remux.GET, path: "/get"}, want: http.StatusForbidden},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(string(tt.args.method), tt.args.path, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		got := response.Code
		if tt.want != got {
			t.Errorf("got %v, want %v", got, tt.want)
		}
	}
}
