package main

import (
	"net/http"
	"time"
)

func (app *application) healthchekHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		}}

	time.Sleep(8 * time.Second)

	err := app.writeJSON(w, http.StatusOK, data, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
