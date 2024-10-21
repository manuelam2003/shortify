package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"text/template"

	"github.com/manuelam2003/shortify/internal/models"
	"github.com/manuelam2003/shortify/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = linkShortenForm{}

	app.render(w, r, http.StatusOK, "home.html", data)
}

type linkShortenForm struct {
	OriginalURL string
	FieldErrors map[string]string
	validator.Validator
}

func (app *application) shortenLink(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := linkShortenForm{
		OriginalURL: r.FormValue("long_url"),
	}

	form.CheckField(
		strings.HasPrefix(form.OriginalURL, "http://") || strings.HasPrefix(form.OriginalURL, "https://"),
		"url",
		"URL must start with http:// or https://",
	)

	form.CheckField(validator.NotBlank(form.OriginalURL), "url", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "home.html", data)
		return
	}

	shortCode := generateShortCode()
	shortenedURL := fmt.Sprintf("http://localhost:4000/%s", shortCode)

	// TODO: add to request body
	expires := 7

	_, err := app.urls.Insert(shortCode, form.OriginalURL, expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "URL successfully shortened!")

	fmt.Fprintf(w, `<div class="alert alert-success mt-4">Shortened URL: <a href="%s">%s</a></div>`, shortenedURL, shortenedURL)
}

var base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode() string {
	length := 6
	shortCode := make([]byte, length)
	for i := range shortCode {
		shortCode[i] = base62Chars[rand.Intn(len(base62Chars))]
	}
	return string(shortCode)
}

func (app *application) shortenView(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	url, err := app.urls.GetByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	referrer := r.Referer()
	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr

	err = app.stats.LogVisit(url.ID, referrer, userAgent, ipAddress)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, url.LongURL, http.StatusSeeOther)

}

func (app *application) urlStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	// Retrieve the URL record by the short code
	url, err := app.urls.GetByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	visitCount, err := app.stats.GetVisitCount(url.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := templateData{
		URL:        url,
		VisitCount: visitCount,
	}

	files := []string{
		"./ui/html/base.html",
		"./ui/html/partials/nav.html",
		"./ui/html/pages/stats.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	// data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.html", data)
}
