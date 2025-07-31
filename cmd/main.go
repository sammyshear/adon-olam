package main

import (
	"log"
	"net/http"

	"github.com/sammyshear/adon-olam/internal/api"
)

func main() {
	mux := api.MuxWithRoutes()

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Panicf("Failed to start http server: %s", err)
	}
}
