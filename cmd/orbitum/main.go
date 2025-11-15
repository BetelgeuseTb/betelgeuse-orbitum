package main

import (
	"log"
	"orbitum/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	if err := application.Run(); err != nil {
		log.Fatal("Failed to run application:", err)
	}
}