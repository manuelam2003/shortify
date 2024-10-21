package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("POST /shorten", app.shortenLink)
	mux.HandleFunc("GET /links/{shortCode}", app.shortenView)
	mux.HandleFunc("GET /links/{shortCode}/stats", app.urlStats)

	mux.HandleFunc("GET /user/signup", app.userSignup)
	// mux.Handle("POST /user/signup", app.userSignupPost)
	// mux.Handle("GET /user/login", app.userLogin)
	// mux.Handle("POST /user/login", app.userLoginPost)

	standard := alice.New(app.recoverPanic, app.logRequest, app.commonHeaders, app.sessionManager.LoadAndSave)

	return standard.Then(mux)
}
