package main

import (
	"log"
	"microblog/httpapi"
)

func main() {
	server := httpapi.NewServer()
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
