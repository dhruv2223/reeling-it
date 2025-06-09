package data

import "github.com/dhruv2223/reeling-it/models"

type MovieStorage interface {
	GetTopMovies() ([]models.Movie, error)
	GetRandomMovies() ([]models.Movie, error)
	GetMovieById(id int) (models.Movie, error)
	SearchMovieByName(name string) ([]models.Movie, error)
	GetAllGenres([]models.Genre, error)
}
