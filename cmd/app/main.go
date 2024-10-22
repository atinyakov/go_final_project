package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/atinyakov/go_final_project/internal"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	envPath := "../../.env"
	if _, err := os.Stat(".env"); err == nil {
		// if not exists in container
		envPath = ".env"
	}

	// load env variables
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db, err := internal.InitDb()
	if err != nil {
		panic(err)
	}

	defer db.Close()
	webDir := "../../web"
	if _, err := os.Stat("web"); err == nil {
		// if not exists in container
		webDir = "web"
	}

	envPort := os.Getenv("TODO_PORT")
	if envPort == "" {
		envPort = "7540"
	}

	fmt.Println("Got a port:", envPort)

	internal.InitServer(webDir, db)

	err = http.ListenAndServe("0.0.0.0:"+envPort, nil)
	fmt.Println("Server started on port=", envPort)

	if err != nil {
		panic(err)
	}
	fmt.Println("Stopping App")
}
