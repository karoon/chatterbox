package main

import (
	"fmt"

	"github.com/pborman/uuid"
)

func main() {
	uuid := uuid.New()

	fmt.Println(uuid)
}
