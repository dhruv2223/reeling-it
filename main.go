package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/dhruv2223/reeling-it/data"
	"github.com/dhruv2223/reeling-it/handlers"
	"github.com/dhruv2223/reeling-it/logger"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return
	}

	dbConnString := os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
		return
	}
	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)

	}
	defer db.Close()

	var addr string = ":8080"
	server := http.NewServeMux()
	movieRepo, _ := data.NewMovieRepository(db, appLogger)

	movieHandler := handlers.NewMovieHandler(movieRepo, appLogger)
	server.HandleFunc("/api/movies/top", movieHandler.GetTopMovies)
	server.HandleFunc("/api/movies/random", movieHandler.GetRandomMovies)
	server.HandleFunc("/api/movies/search", movieHandler.SearchMovies)
	server.HandleFunc("/api/movies", movieHandler.GetMovie)
	server.HandleFunc("/api/genres", movieHandler.GetGenre)

	catchAllClientRoutesHandler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	}
	server.HandleFunc("/movies", catchAllClientRoutesHandler)

	server.HandleFunc("/movies/", catchAllClientRoutesHandler)
	server.HandleFunc("/account", catchAllClientRoutesHandler)

	server.Handle("/", http.FileServer(http.Dir("./public")))

	err = http.ListenAndServe(addr, server)
	if err != nil {
		appLogger.Error("Error starting server", err)
	}
}
