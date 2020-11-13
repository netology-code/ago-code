package identificator

import (
	"bytes"
	"github.com/go-chi/chi"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIdentificatorHTTPMux(t *testing.T) {
	mux := http.NewServeMux()
	identificatorMd := Identificator
	mux.Handle("/get",
		identificatorMd(http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				identifier, err := Identifier(request.Context())
				if err != nil {
					t.Fatal(err)
				}
				_, err = writer.Write([]byte(*identifier))
				if err != nil {
					t.Fatal(err)
				}
			})),
	)

	type args struct {
		method string
		path   string
		addr   string
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "GET", args: args{method: "GET", path: "/get", addr: "192.0.2.1:12345"}, want: []byte("192.0.2.1")},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		request.RemoteAddr = tt.args.addr

		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}

func TestIdentificatorChi(t *testing.T) {
	router := chi.NewRouter()
	identificatorMd := Identificator
	router.With(identificatorMd).Get(
		"/get",
		func(writer http.ResponseWriter, request *http.Request) {
			identifier, err := Identifier(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			_, err = writer.Write([]byte(*identifier))
			if err != nil {
				t.Fatal(err)
			}
		},
	)

	type args struct {
		method string
		path   string
		addr   string
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "GET", args: args{method: "GET", path: "/get", addr: "192.0.2.1:12345"}, want: []byte("192.0.2.1")},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		request.RemoteAddr = tt.args.addr

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}
