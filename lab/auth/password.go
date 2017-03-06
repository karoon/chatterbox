package main

import (
	"chatterbox/mqtt/auth"
	"log"
)

func main() {
	auth.SetPassword("vahid", "1")
	log.Println(auth.CheckAuth("vahid", "1"))
}
