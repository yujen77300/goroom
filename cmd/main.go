package main

import (
	"log"
	"github.com/yujen77300/goroom/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err.Error())
	}
}