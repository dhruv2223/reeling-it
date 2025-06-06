package main

import (
	"log"
	"net/http"

	"github.com/dhruv2223/reeling-it/logger"
)

func InitializeLogger() *logger.Logger {
	appLogger, err := logger.NewLogger("./logs/app.log")
	if err != nil {
		log.Fatalf("Could not create logger: %v", err)
	}
	return appLogger
}

func main() {
	var addr string = ":8080"
	server := http.NewServeMux()
	server.Handle("/", http.FileServer(http.Dir("./public")))
	appLogger := InitializeLogger()
	defer appLogger.Close()
	err := http.ListenAndServe(addr, server)
	if err != nil {
		appLogger.Error("Error starting server", err)
	}

}
