package main

import (
	"encoding/base64"
	"io/ioutil"
	"lectionjwt/pkg/jwt/symmetric"
	"log"
	"os"
	"time"
)

func main() {
	if err := execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

type Data struct {
	UserID int64    `json:"userId"`
	Roles  []string `json:"roles"`
	Issued int64    `json:"iat"`
	Expire int64    `json:"exp"`
}

func execute() error {
	secretKeyBytes, err := ioutil.ReadFile("keys/symmetric.key")
	if err != nil {
		return err
	}
	secretKey, err := base64.StdEncoding.DecodeString(string(secretKeyBytes))
	if err != nil {
		return err
	}

	data := &Data{
		UserID: 1,
		Roles:  []string{"ADMIN"},
		Issued: time.Now().Unix(),
		Expire: time.Now().Add(time.Minute * 10).Unix(),
	}

	token, err := symmetric.Encode(data, secretKey)
	if err != nil {
		return err
	}

	log.Printf("generated token: %s", token)

	verified, err := symmetric.Verify(token, secretKey)
	if err != nil {
		return err
	}
	log.Printf("verification result: %t", verified)

	var decoded *Data
	err = symmetric.Decode(token, &decoded)
	if err != nil {
		return err
	}
	log.Printf("decoded token: %#v", decoded)

	return nil
}
