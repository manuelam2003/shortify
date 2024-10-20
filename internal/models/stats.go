package models

import (
	"database/sql"
	"time"
)

type Stats struct {
	ID        int
	URLID     int
	ClickTime time.Time
	Referrer  string
	UserAgent string
	IPAddress string
}

type StatsModel struct {
	DB *sql.DB
}

func (m *StatsModel) LogVisit(urlID int, referrer, userAgent, ipAddress string) error {
	query := `
        INSERT INTO url_analytics (url_id, referrer, user_agent, ip_address)
        VALUES (?, ?, ?, ?)
    `
	_, err := m.DB.Exec(query, urlID, referrer, userAgent, ipAddress)
	return err
}

func (m *StatsModel) GetVisitCount(urlID int) (int, error) {
	query := `SELECT COUNT(*) FROM url_analytics WHERE url_id = ?`
	var visitCount int
	err := m.DB.QueryRow(query, urlID).Scan(&visitCount)
	if err != nil {
		return 0, err
	}
	return visitCount, nil
}
