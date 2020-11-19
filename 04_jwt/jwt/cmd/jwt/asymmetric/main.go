package main

import (
	"io/ioutil"
	"lectionjwt/pkg/jwt/asymmetric"
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
	publicKey, err := ioutil.ReadFile("./keys/public.key")
	if err != nil {
		return err
	}

	privateKey, err := ioutil.ReadFile("./keys/private.key")
	if err != nil {
		return err
	}

	data := &Data{
		UserID: 1,
		Roles:  []string{"ADMIN"},
		Issued: time.Now().Unix(),
		Expire: time.Now().Add(time.Minute * 10).Unix(),
	}

	token, err := asymmetric.Encode(data, privateKey)
	if err != nil {
		return err
	}

	log.Printf("generated token: %s", token)

	verified, err := asymmetric.Verify(token, publicKey)
	if err != nil {
		return err
	}
	log.Printf("verification result: %t", verified)

	var decoded *Data
	err = asymmetric.Decode(token, &decoded)
	if err != nil {
		return err
	}
	log.Printf("decoded token: %#v", decoded)

	return nil
}
