package main

import (
	"log"
	"urlshortener/internal/config"
)

func main() {
	cfg := config.MustLoad()
	log.Println(cfg)
}
