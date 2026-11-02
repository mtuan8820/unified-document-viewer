package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /service/records", listRecords) // VIN comes as ?vin=...

	srv := &http.Server{
		Addr:           ":8082",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("service mock listening on :8082")
	log.Fatal(srv.ListenAndServe())
}
