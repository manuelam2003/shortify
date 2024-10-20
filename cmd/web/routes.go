package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("POST /shorten", app.shortenLink)
	mux.HandleFunc("GET /{shortCode}", app.shortenView)
	mux.HandleFunc("GET /{shortCode}/stats", app.urlStats)

	mux.HandleFunc("GET /user/signup", app.userSignup)
	// mux.Handle("POST /user/signup", app.userSignupPost)
	// mux.Handle("GET /user/login", app.userLogin)
	// mux.Handle("POST /user/login", app.userLoginPost)

	return mux
}
