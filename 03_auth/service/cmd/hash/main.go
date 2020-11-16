package main

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		return
	}

	// hash будет каждый раз разным для одних и тех же данных - это нормально (т.к. соль разная)
	log.Print(string(hash))
}
