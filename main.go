package main

import (
	"log"
	"microblog/httpapi"
	"microblog/storage"
	"microblog/storage/inmemory"
	"microblog/storage/mongo"
	"os"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	port := getEnv("SERVER_PORT", "8080")
	address := "0.0.0.0:" + port
	storageMode := getEnv("STORAGE_MODE", "inmemory")
	var storageVar storage.Storage
	if storageMode == "inmemory" {
		storageVar = inmemory.NewStorage()
	} else {
		mongoUrl := getEnv("MONGO_URL", "mongodb://localhost:27017")
		mongoDbName := getEnv("MONGO_DBNAME", "microblog")
		storageVar = mongo.NewStorage(mongoUrl, mongoDbName)
	}
	server := httpapi.NewServer(storageVar, address)
	log.Printf("Storage mode %s", storageMode)
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
