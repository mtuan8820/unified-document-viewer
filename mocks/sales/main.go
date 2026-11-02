package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/vehicles/{vin}/documents", &MyHandler{})
	srv := &http.Server{
		Addr:           ":8081",
		Handler:        mux,
		ReadTimeout:    10 * time.Second, // Maximum duration for reading the entire request
		WriteTimeout:   10 * time.Second, // Maximum duration before timing out writes of the response
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(srv.ListenAndServe())
}
