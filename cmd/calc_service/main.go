package main

import (
	"log"
	"net/http"
	"os"

	"calc_service/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := server.NewServer()
	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, srv.Router))
}
