package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	
	data := envelop{
		"status": "available",
		"environment": app.config.env,
		"version": version,
	}

	err := app.writeJson(w, http.StatusOK, data, nil)

	if err != nil {
		app.logger.Println(err)
		http.Error(w, "the server ecountered a problem and could not process your request ", http.StatusInternalServerError)
		return
	}

}

