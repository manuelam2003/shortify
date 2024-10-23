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

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// Update the handler so it displays the signup page.
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	fmt.Println("sjcnvjntjtnrbtjn")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Add the ID of the current user to the session, so that they are now
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
