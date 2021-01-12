package authenticator

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"github.com/netology-code/remux/pkg/remux"
	"testing"
)

func TestAuthenticator(t *testing.T) {
	mux := remux.NewReMux()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "USERPROFILE", nil
	})
	if err := mux.RegisterPlain(
		remux.GET,
		"/get",
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			data := profile.(string)

			writer.Write([]byte(data))
		}),
		authenticatorMd,
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
		want []byte
	}{
		{name: "GET", args: args{method: remux.GET, path: "/get"}, want: []byte("USERPROFILE")},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(string(tt.args.method), tt.args.path, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}
