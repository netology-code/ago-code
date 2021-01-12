package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
)

const defaultPort = "9999"
const defaultHost = "0.0.0.0"
const defaultCertificatePath = "./tls/certificate.pem"
const defaultPrivateKeyPath = "./tls/key.pem"

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	certificatePath, ok := os.LookupEnv("APP_CERT_PATH")
	if !ok {
		certificatePath = defaultCertificatePath
	}

	privateKeyPath, ok := os.LookupEnv("APP_PRIVATE_KEY_PATH")
	if !ok {
		privateKeyPath = defaultPrivateKeyPath
	}

	if err := execute(net.JoinHostPort(host, port), certificatePath, privateKeyPath); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string, certificatePath string, privateKeyPath string) (err error) {
	return http.ListenAndServeTLS(addr, certificatePath, privateKeyPath, &handler{});
}

type ResponseDTO struct {
	Status string `json:"status"`
}

type handler struct {}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	resp := &ResponseDTO{Status: "ok"}
	respBody, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	_, err = writer.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}

