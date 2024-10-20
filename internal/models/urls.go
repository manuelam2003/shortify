package models

import (
	"database/sql"
	"errors"
	"time"
)

type URL struct {
	ID        int
	ShortCode string
	LongURL   string
	// UserID int
	ExpiresAt time.Time
	CreatedAt time.Time
}

type URLModel struct {
	DB *sql.DB
}

func (m *URLModel) Insert(shortURL, longURL string, expires int) (int, error) {
	stmt := `
		INSERT INTO urls (short_code, long_url, expiration) 
		VALUES (?, ?, ?)
	`

	// Calculate expiration date if 'expires' is provided
	var expiration sql.NullTime
	if expires > 0 {
		expiration.Valid = true
		expiration.Time = time.Now().AddDate(0, 0, expires) // Add 'expires' days from now
	} else {
		expiration.Valid = false // No expiration
	}

	// Execute the insert query
	result, err := m.DB.Exec(stmt, shortURL, longURL, expiration)
	if err != nil {
		return 0, err
	}

	// Get the last inserted ID (primary key)
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *URLModel) Get(id int) (URL, error) {
	// SQL query to select the URL by ID
	stmt := `
		SELECT id, short_code, long_url, expiration, created_at
		FROM urls
		WHERE id = ?
	`

	// Create a URL instance to store the result
	var url URL
	var expiration sql.NullTime

	// Execute the query
	err := m.DB.QueryRow(stmt, id).Scan(
		&url.ID, &url.ShortCode, &url.LongURL, &expiration, &url.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URL{}, errors.New("url not found")
		}
		return URL{}, err
	}

	// If expiration is valid, set it, otherwise leave it at zero value
	if expiration.Valid {
		url.ExpiresAt = expiration.Time
	}

	return url, nil
}

func (m *URLModel) GetByShortCode(shortCode string) (URL, error) {
	stmt := `
		SELECT id, short_code, long_url, expiration, created_at
		FROM urls
		WHERE short_code = ?`

	var url URL
	var expiration sql.NullTime

	err := m.DB.QueryRow(stmt, shortCode).Scan(
		&url.ID, &url.ShortCode, &url.LongURL, &expiration, &url.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URL{}, errors.New("url not found")
		}
		return URL{}, err
	}

	if expiration.Valid {
		url.ExpiresAt = expiration.Time
	}

	return url, nil
}
