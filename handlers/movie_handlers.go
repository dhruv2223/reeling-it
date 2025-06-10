package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dhruv2223/reeling-it/data"
	"github.com/dhruv2223/reeling-it/logger"
	"github.com/dhruv2223/reeling-it/models"
)

type MovieHandler struct {
	storage data.MovieStorage
	logger  *logger.Logger
}

func NewMovieHandler(storage data.MovieStorage, logger *logger.Logger) *MovieHandler {
	return &MovieHandler{
		storage: storage,
		logger:  logger,
	}
}
func (mh *MovieHandler) writeJSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		mh.logger.Error("Failed to encode data", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}
	return nil
}
func (mh *MovieHandler) GetTopMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := mh.storage.GetTopMovies()
	if err != nil {
		mh.logger.Error("Failed to respond with top movies", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = mh.writeJSONResponse(w, movies)
	if err == nil {
		mh.logger.Info("Succesfully served top movies")
		return
	}

}
func (mh *MovieHandler) GetRandomMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := mh.storage.GetRandomMovies()
	if err != nil {
		mh.logger.Error("Failed to respond with random movies", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = mh.writeJSONResponse(w, movies)
	if err == nil {
		mh.logger.Info("Succesfully served random movies")
		return
	}
}

func (mh *MovieHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query()["id"]
	if id != nil {
		finalId, err := strconv.Atoi(id[0])
		if err != nil {
			mh.logger.Error(fmt.Sprintf("Failed to parse the id: %s", id), err)
			http.Error(w, "Invalid Id", http.StatusBadRequest)
		}
		movie, err := mh.storage.GetMovieByID(finalId)
		if err != nil {
			mh.logger.Error(fmt.Sprintf("Failed to get movie for id: %v", finalId), err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		err = mh.writeJSONResponse(w, movie)

		if err == nil {
			mh.logger.Info("Succesfully responded with all genres")
		}
		//need to do proper error management here as i am returning internal server error even if
		//the user is sending wrong id
		// need to create a func that checks for the error and then returns appropriate error message
	}
}
func (mh *MovieHandler) GetGenre(w http.ResponseWriter, r *http.Request) {
	genres, err := mh.storage.GetAllGenres()
	if err != nil {
		mh.logger.Error("Failed to send all genres", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = mh.writeJSONResponse(w, genres)
	if err == nil {
		mh.logger.Info("Succesfully responded with all genres")
	}

}
func (mh *MovieHandler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	order := r.URL.Query().Get("order")
	genreStr := r.URL.Query().Get("genre")
	var genre *int
	if genreStr != "" {
		genreId, err := strconv.Atoi(genreStr)
		if err != nil {
			mh.logger.Error("Failed to parse genre ID: "+genreStr, err)
			http.Error(w, "Invalid Genre ID", http.StatusBadRequest)
			return
		}
		genre = &genreId
	}
	var movies []models.Movie
	movies, err := mh.storage.SearchMoviesByName(query, order, genre)
	if err != nil {
		mh.logger.Error("Failed to search movies", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = mh.writeJSONResponse(w, movies)
	if err == nil {
		mh.logger.Info("Succesfully responded with searched movies")
		return
	}

}
