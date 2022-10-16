package main

import (
	"log"
	"microblog/httpapi"
	"microblog/storage/inmemory"
)

func main() {
	storage := inmemory.NewStorage()
	server := httpapi.NewServer(storage)
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
