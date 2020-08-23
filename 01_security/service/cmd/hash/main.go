package main

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

func main() {
	// Пароли нельзя хранить в открытом виде, т.к. большинство пользователей использует одни и те же пароли не всех сервисах.
	// Поэтому их хэшируют - md5, sha1 уже считаются небезопасными, поэтому используем bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		os.Exit(-1)
		return
	}

	// hash будет каждый раз разным для одних и тех же данных - это нормально
	log.Print(string(hash))
}
