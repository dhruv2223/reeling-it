package main

import (
	"log"
	"net/http"

	"github.com/dhruv2223/reeling-it/handlers"
	"github.com/dhruv2223/reeling-it/logger"
)

func InitializeLogger() *logger.Logger {
	appLogger, err := logger.NewLogger("./app.log")
	if err != nil {
		log.Fatalf("Could not create logger: %v", err)
	}
	return appLogger
}

func main() {

	appLogger := InitializeLogger()
	defer appLogger.Close()
	var addr string = ":8080"
	server := http.NewServeMux()

	movieHandler := handlers.MovieHandler{}
	server.HandleFunc("/api/movies/top", movieHandler.GetTopMovies)
	server.HandleFunc("/api/movies/random", movieHandler.GetRandomMovies)
	server.Handle("/", http.FileServer(http.Dir("./public")))
	err := http.ListenAndServe(addr, server)
	if err != nil {
		appLogger.Error("Error starting server", err)
	}
}
