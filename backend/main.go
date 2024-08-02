package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sameer-gits/CMS/database"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Error loading .env file")
	}
}

func main() {
	defer func() {
		database.RedisClose()
		database.DbClose()
	}()

	err := database.DbInit(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	err = database.RedisInit(os.Getenv("REDIS_URL_0"), os.Getenv("REDIS_URL_1"))
	if err != nil {
		log.Fatalf("Redis initialization failed: %v", err)
	}

	routes()
}
