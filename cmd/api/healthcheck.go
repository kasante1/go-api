package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	
	data := envelope{
		"status": "available",
		"environment": app.config.env,
		"version": version,
	}

	err := app.writeJson(w, http.StatusOK, data, nil)

	if err != nil {
		app.logger.Println(err)
		app.serverErrorResponse(w, r, err)
		return
	}

}

