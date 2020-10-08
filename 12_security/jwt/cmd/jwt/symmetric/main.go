package main

import (
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
	secretKey := []byte("top secret")

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
