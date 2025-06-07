package models

type Movie struct {
	ID          int
	TMBD_ID     int
	Title       string
	TagLine     string
	ReleaseYear int
	Genres      []Genre
	Overview    *string
	Score       *float64
	Popularity  *float64
	Keywords    []string
	Language    *string
	PosterURL   *string
	TrailerURL  *string
	Casting     []Actor
}
