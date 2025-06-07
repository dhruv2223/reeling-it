package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dhruv2223/reeling-it/models"
)

type MovieHandler struct{}

func (mh *MovieHandler) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (mh *MovieHandler) GetTopMovies(w http.ResponseWriter, r *http.Request) {
	movies := []models.Movie{
		{
			ID:          1,
			TMBD_ID:     181,
			Title:       "The Dark Knight",
			ReleaseYear: 2002,
			Genres:      []models.Genre{{ID: 1, Name: "Action"}, {ID: 2, Name: "Crime"}},
			Keywords:    []string{},
			Casting:     []models.Actor{{ID: 1, FirstName: "Dhruv", LastName: "Chotalia"}},
		},

		{
			ID:          1,
			TMBD_ID:     181,
			Title:       "The Light Morning",
			ReleaseYear: 2002,
			Genres:      []models.Genre{{ID: 1, Name: "Action"}, {ID: 2, Name: "Love"}},
			Keywords:    []string{},
			Casting:     []models.Actor{{ID: 1, FirstName: "Dhruv", LastName: "Chotalia"}},
		},
	}
	mh.writeJSONResponse(w, movies)
}

func (mh *MovieHandler) GetRandomMovies(w http.ResponseWriter, r *http.Request) {
	movies := []models.Movie{
		{
			ID:          1,
			TMBD_ID:     181,
			Title:       "The Dark Knight",
			ReleaseYear: 2002,
			Genres:      []models.Genre{{ID: 1, Name: "Action"}, {ID: 2, Name: "Crime"}},
			Keywords:    []string{},
			Casting:     []models.Actor{{ID: 1, FirstName: "Dhruv", LastName: "Chotalia"}},
		},

		{
			ID:          1,
			TMBD_ID:     181,
			Title:       "The Light Morning",
			ReleaseYear: 2002,
			Genres:      []models.Genre{{ID: 1, Name: "Action"}, {ID: 2, Name: "Love"}},
			Keywords:    []string{},
			Casting:     []models.Actor{{ID: 1, FirstName: "Dhruv", LastName: "Chotalia"}},
		},
	}
	mh.writeJSONResponse(w, movies)
}
