package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}

func migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS trips (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT,
			start_time TEXT NOT NULL DEFAULT '10:00',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_trips_user_id ON trips(user_id)`,
		`CREATE TABLE IF NOT EXISTS trip_places (
			id TEXT PRIMARY KEY,
			trip_id TEXT NOT NULL,
			place_id TEXT NOT NULL,
			name TEXT NOT NULL,
			lat REAL NOT NULL,
			lng REAL NOT NULL,
			kind TEXT NOT NULL,
			stay_minutes INTEGER NOT NULL DEFAULT 60,
			order_index INTEGER NOT NULL,
			reason TEXT,
			review_summary TEXT,
			photo_url TEXT,
			FOREIGN KEY (trip_id) REFERENCES trips(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_trip_places_trip_id ON trip_places(trip_id)`,
		`CREATE TABLE IF NOT EXISTS shares (
			share_id TEXT PRIMARY KEY,
			trip_id TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (trip_id) REFERENCES trips(id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

