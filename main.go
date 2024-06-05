package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/sameer-gits/CMS/database"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.InitDB()
	defer database.Conn.Close(context.Background())
	// err = createSchema(conn)
	// if err != nil {
	//     log.Fatalf("Unable to create schema: %v\n", err)
	// }

	routes()
}
