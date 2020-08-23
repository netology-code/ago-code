package identificator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"github.com/netology-code/remux/pkg/remux"
	"testing"
)

func TestIdentificator(t *testing.T) {
	mux := remux.NewReMux()
	identificatorMd := Identificator
	if err := mux.RegisterPlain(
		remux.GET,
		"/get",
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			identifier, err := Identifier(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			writer.Write([]byte(*identifier))
		}),
		identificatorMd,
	); err != nil {
		t.Fatal(err)
	}

	type args struct {
		method remux.Method
		path   string
		addr   string
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "GET", args: args{method: remux.GET, path: "/get", addr: "192.0.2.1:12345"}, want: []byte("192.0.2.1")},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(string(tt.args.method), tt.args.path, nil)
		request.RemoteAddr = tt.args.addr

		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}
