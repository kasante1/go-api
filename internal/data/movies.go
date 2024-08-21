package data

import (
	"time"
	"github.com/kasante1/go-api/internal/validator"
	"database/sql"
	"github.com/lib/pq" 
)


type Movie struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title string `json:"title"`
	Year int32 `json:"year,omitempty"`
	Runtime Runtime `json:"runtime,omitempty"`
	Genres []string `json:"genre,omitempty"`
	Version	int32 `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie){
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500 , "title", "the title must be less than 500 characters")

	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must have at least one genre")
	v.Check(len(movie.Genres) >= 1, "genres", "must not have more than 5 genres")

	v.Check(validator.Unique(movie.Genres), "genres", "must not have duplicate genres")
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
	}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

func (m MovieModel) Update(movie *Movie) error {
	return nil
}

func (m MovieModel) Delete(id int64) error {
	return nil
}