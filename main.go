package main

import (
	"github.com/go-redis/redis/v8"
	"log"
	"microblog/httpapi"
	"microblog/storage"
	"microblog/storage/cached"
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
	storageMode := getEnv("STORAGE_MODE", "cached")
	var storageVar storage.Storage
	if storageMode == "inmemory" {
		storageVar = inmemory.NewStorage()
	} else if storageMode == "mongo" {
		mongoUrl := getEnv("MONGO_URL", "mongodb://localhost:27017")
		mongoDbName := getEnv("MONGO_DBNAME", "microblog")
		storageVar = mongo.NewStorage(mongoUrl, mongoDbName)
	} else if storageMode == "cached" {
		mongoUrl := getEnv("MONGO_URL", "mongodb://localhost:27017")
		mongoDbName := getEnv("MONGO_DBNAME", "microblog")
		mongoStorage := mongo.NewStorage(mongoUrl, mongoDbName)
		redisUrl := getEnv("REDIS_URL", "127.0.0.1:6379")
		redisClient := redis.NewClient(&redis.Options{Addr: redisUrl})
		storageVar = cached.NewStorage(redisClient, mongoStorage)
	} else {
		log.Printf("Unknown mode")
		return
	}
	server := httpapi.NewServer(storageVar, address)
	log.Printf("Storage mode %s", storageMode)
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
