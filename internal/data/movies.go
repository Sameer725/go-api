package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"movies.samkha.net/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // - directive omits the item from json
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`    // - omitempty omits the item if empty/falsy value
	Runtime   Runtime   `json:"runtime,omitempty"` // - string directive changes the field item to string
	Genres    []string  `json:",omitempty"`        // leaving 1st directive blank leave the filed title as it is
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *sql.DB
}

func (model MovieModel) Insert(movie *Movie) error {
	query := `INSERT INTO movies(title,year,runtime,genres) VALUES($1,$2,$3,$4) RETURNING id,created_at, version`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return model.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (model MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id,created_at,title,year,runtime,version,genres FROM movies WHERE id=$1`
	//to demo timeout
	// query := `SELECT pg_sleep(7), id,created_at,title,year,runtime,version,genres FROM movies WHERE id=$1`

	var movie Movie

	//3 sec timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//cancel the context before the GET returns
	defer cancel()

	err := model.DB.QueryRowContext(ctx, query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, &movie.Version, pq.Array(&movie.Genres))
	//passing the context with timeout, terminates the long running query if it taken more that defined timeout
	// err := model.DB.QueryRowContext(ctx, query, id).Scan(&[]byte{}, &movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, &movie.Version, pq.Array(&movie.Genres))

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (model MovieModel) Update(movie *Movie) error {
	query := `UPDATE movies SET title=$1,year=$2,runtime=$3,genres=$4,version=version+1 WHERE id=$5 AND version=$6 RETURNING version`
	args := []any{&movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.ID, &movie.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := model.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (model MovieModel) Delete(id int64) error {
	query := `DELETE FROM movies WHERE id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := model.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
