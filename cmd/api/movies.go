package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kasante1/go-api/internal/data"
	"github.com/kasante1/go-api/internal/validator"
)


func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string 	`json:"title"`
		Year int32 		`json:"year"`
		Runtime data.Runtime `json:"runtime"` 
		Genres []string `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	v.Check(input.Title != "", "title", "must be provided")
	v.Check(len(input.Title) <= 500 , "title", "the title must be less than 500 characters")

	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(input.Genres != nil, "genres", "must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must have at least one genre")
	v.Check(len(input.Genres) >= 1, "genres", "must not have more than 5 genres")

	v.Check(validator.Unique(input.Genres), "genres", "must not have duplicate genres")
	if !v.Valid(){
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil{
	app.notFoundResponse(w, r)
	return
	}

	movie := data.Movie{
		ID: id,
		CreatedAt: time.Now(),
		Title: "Casablanca",
		Runtime: 102,
		Genres: []string{"drama", "romance", "war"},
		Version: 1,
	}



	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.logger.Println(err)
		app.serverErrorResponse(w, r, err)
	}
}
