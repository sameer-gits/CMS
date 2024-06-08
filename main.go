package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

const (
	serverCode   = http.StatusInternalServerError
	unauthorized = http.StatusUnauthorized
	statusOK     = http.StatusOK
	badCode      = http.StatusBadRequest
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	InitDB()
	defer Conn.Close(context.Background())
	// err = createSchema(conn)
	// if err != nil {
	//     log.Fatalf("Unable to create schema: %v\n", err)
	// }

	routes()
}
