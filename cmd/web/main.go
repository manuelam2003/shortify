package main

import (
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/manuelam2003/shortify/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	logger         *slog.Logger
	urls           *models.URLModel
	stats          *models.StatsModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:         logger,
		urls:           &models.URLModel{DB: db},
		stats:          &models.StatsModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", "addr", srv.Addr)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./shortify.db")
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	createTables := `
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			short_code TEXT UNIQUE NOT NULL,
			long_url TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expiration DATETIME
		);
	
		CREATE TABLE IF NOT EXISTS url_analytics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url_id INTEGER NOT NULL,
			click_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			referrer TEXT,
			user_agent TEXT,
			ip_address TEXT,
			FOREIGN KEY(url_id) REFERENCES urls(id) ON DELETE CASCADE
		);
	
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	_, err = db.Exec(createTables)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
