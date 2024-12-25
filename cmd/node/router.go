package main

import (
	"net/http"
)

func (app *App) getRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /transactions", app.createTransaction)

	return mux
}
