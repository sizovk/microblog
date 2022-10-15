package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloWorld(rw http.ResponseWriter, r *http.Request) {
	name := "stranger"
	if customName := r.URL.Query().Get("name"); customName != "" {
		name = customName
	}
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(fmt.Sprintf("Hello, %s!\n", name)))
}

func main() {
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: http.HandlerFunc(helloWorld),
	}
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
