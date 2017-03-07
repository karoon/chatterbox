package main

import (
	"log"

	"golang.org/x/crypto/scrypt"
)

func main() {
	dk, err := scrypt.Key([]byte("vahid"), []byte(""), 16384, 8, 1, 32)
	log.Println(string(dk), err)
}
