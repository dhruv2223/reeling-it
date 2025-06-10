package data

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/dhruv2223/reeling-it/logger"
	"github.com/dhruv2223/reeling-it/models"
)

var (
	ErrMovieNotFound = errors.New("movie not found")
)

type MovieRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewMovieRepository(db *sql.DB, logger *logger.Logger) (*MovieRepository, error) {
	return &MovieRepository{
		db:     db,
		logger: logger,
	}, nil
}

const defaultLimit int = 20

func (r *MovieRepository) GetTopMovies() ([]models.Movie, error) {
	query := ` 
  SELECT id,tmdb_id,title,tagline, release_year,  overview, score, popularity, language, poster_url, trailer_url 
  FROM movies 
  ORDER BY popularity DESC  
  LIMIT $1
  `
	return r.getMovies(query)
}

func (r *MovieRepository) GetRandomMovies() ([]models.Movie, error) {
	query := ` 
  SELECT id,tmdb_id,title,tagline, release_year,  overview, score, popularity, language, poster_url, trailer_url 
  FROM movies 
  ORDER BY random() DESC  
  LIMIT $1
  `
	return r.getMovies(query)
}

func (r *MovieRepository) getMovies(query string) ([]models.Movie, error) {
	rows, err := r.db.Query(query, defaultLimit)
	if err != nil {
		r.logger.Error("Failed to query movies", err)
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(
			&m.ID, &m.TMDB_ID, &m.Title, &m.Tagline, &m.ReleaseYear,
			&m.Overview, &m.Score, &m.Popularity, &m.Language,
			&m.PosterURL, &m.TrailerURL,
		); err != nil {
			r.logger.Error("Failed to scan movie row", err)
			return nil, err
		}
		movies = append(movies, m)
	}

	return movies, nil
}

func (r *MovieRepository) GetMovieByID(id int) (models.Movie, error) {
	query := ` 
  SELECT id,tmdb_id,title,tagline, release_year,  overview, score, popularity, language, poster_url, trailer_url 
  FROM movies  
  WHERE id = $1
  `
	row := r.db.QueryRow(query, id)
	var m models.Movie
	err := row.Scan(
		&m.ID, &m.TMDB_ID, &m.Title, &m.Tagline, &m.ReleaseYear,
		&m.Overview, &m.Score, &m.Popularity, &m.Language,
		&m.PosterURL, &m.TrailerURL,
	)
	if err == sql.ErrNoRows {
		r.logger.Error("Movie not found", ErrMovieNotFound)
		return models.Movie{}, ErrMovieNotFound
	}
	if err != nil {
		r.logger.Error("Failed to query by movie id", err)
		return models.Movie{}, err
	}
	err = r.fetchMovieRelations(&m)
	if err != nil {
		r.logger.Error("Failed to fetch movie relations for movie "+strconv.Itoa(m.ID), err)
		return models.Movie{}, err
	}

	return m, nil
	// Implementation for fetching a movie by ID from the database
}

func (r *MovieRepository) SearchMoviesByName(name string, order string, genre *int) ([]models.Movie, error) {
	orderBy := "popularity DESC"
	switch order {
	case "score":
		orderBy = "score DESC"
	case "name":
		orderBy = "title"
	case "date":
		orderBy = "release_year DESC"
	}

	genreFilter := ""
	if genre != nil {
		genreFilter = ` AND ((SELECT COUNT(*) FROM movie_genres 
								WHERE movie_id=movies.id 
								AND genre_id=` + strconv.Itoa(*genre) + `) = 1) `
	}

	// Fetch movies by name
	query := `
		SELECT id, tmdb_id, title, tagline, release_year, overview, score, 
		       popularity, language, poster_url, trailer_url
		FROM movies
		WHERE (title ILIKE $1 OR overview ILIKE $1) ` + genreFilter + `
		ORDER BY ` + orderBy + `
		LIMIT $2
	`
	rows, err := r.db.Query(query, "%"+name+"%", defaultLimit)
	if err != nil {
		r.logger.Error("Failed to search movies by name", err)
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(
			&m.ID, &m.TMDB_ID, &m.Title, &m.Tagline, &m.ReleaseYear,
			&m.Overview, &m.Score, &m.Popularity, &m.Language,
			&m.PosterURL, &m.TrailerURL,
		); err != nil {
			r.logger.Error("Failed to scan movie row", err)
			return nil, err
		}
		movies = append(movies, m)
	}

	return movies, nil
}

func (r *MovieRepository) GetAllGenres() ([]models.Genre, error) {
	query := `SELECT id, name FROM genres ORDER BY id`
	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error("Failed to query all genres", err)
		return nil, err
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var g models.Genre
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			r.logger.Error("Failed to scan genre row", err)
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func (r *MovieRepository) fetchMovieRelations(m *models.Movie) error {
	// Implementation for fetching movie relations like casting, keywords, etc.
	genreQuery := ` 
  SELECT g.id,g.name 
  FROM genres g 
  INNER JOIN 
  movie_genres mg  
  ON 
  g.id = mg.genre_id  
  WHERE 
  mg.movie_id = $1 
  `
	genreRows, err := r.db.Query(genreQuery, m.ID)
	if err != nil {
		r.logger.Error("Failed to fetch movie genres for movie "+strconv.Itoa(m.ID), err)
		return err
	}
	defer genreRows.Close()
	for genreRows.Next() {
		var g models.Genre
		err := genreRows.Scan(&g.ID, &g.Name)
		if err != nil {
			r.logger.Error("Failed to scan movie genre row "+strconv.Itoa(m.ID), err)
			return err
		}
		m.Genres = append(m.Genres, g)
	}
	keywordsQuery := `
  SELECT k.word From keywords k 
  INNER JOIN movie_keywords mk 
  ON k.id = mk.keyword_id 
  WHERE 
  mk.movie_id = $1 
  `
	keyWordRows, err := r.db.Query(keywordsQuery, m.ID)
	if err != nil {
		r.logger.Error("Failed to fetch movie keywords for movie "+strconv.Itoa(m.ID), err)
		return err
	}
	defer keyWordRows.Close()
	for keyWordRows.Next() {
		var keyword string
		err := keyWordRows.Scan(&keyword)
		if err != nil {
			r.logger.Error("Failed to scan movie keyword row "+strconv.Itoa(m.ID), err)
			return err
		}
		m.Keywords = append(m.Keywords, keyword)
	}

	actorsQuery := `
  SELECT a.id,a.first_name,a.last_name,a.image_url 
  FROM actors a 
  INNER JOIN  
  movie_cast mc  
  ON 
  a.id = mc.actor_id  
  WHERE 
  mc.movie_id = $1 
  `
	actorRows, err := r.db.Query(actorsQuery, m.ID)
	if err != nil {
		r.logger.Error("Failed to fetch movie actors for movie "+strconv.Itoa(m.ID), err)
		return err
	}
	for actorRows.Next() {
		var a models.Actor
		err := actorRows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.ImageURL)
		if err != nil {
			r.logger.Error("Failed to scan movie actor row "+strconv.Itoa(m.ID), err)
		}
		m.Casting = append(m.Casting, a)
	}

	return nil
}
