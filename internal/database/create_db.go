package database

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var schema = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY NOT NULL,
	username TEXT NOT NULL UNIQUE,
	hash BLOB NOT NULL,
	is_admin INT NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
	id TEXT PRIMARY KEY NOT NULL,
	name TEXT NOT NULL UNIQUE,
	owner_id TEXT NOT NULL,
	FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS store (
	plugin_id TEXT NOT NULL,
	key TEXT NOT NULL,
	value BLOB NOT NULL,
	PRIMARY KEY (plugin_id, key)
);`

func SetupDB(l hclog.Logger) (*sqlx.DB, error) {
	// Open and ping database
	db, err := sqlx.Connect("sqlite", "db.sqlite")
	if err != nil {
		return nil, err
	}

	// Check if database already exists
	table := ""
	err = db.Get(&table, "SELECT name FROM sqlite_master WHERE type='table' AND name='users'")

	if err == nil {
		// Database has already been created
		l.Debug("database already exists")
		return db, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Unexpected error
		return nil, err
	}

	l.Debug("database does not exist, creating it now")

	// Create schema
	result, err := db.Exec(schema)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	l.Debug("created database schema", "rowsAffected", rows)

	// Create default "root" user
	hash, err := bcrypt.GenerateFromPassword([]byte("changeme!!"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("INSERT INTO users VALUES ($1, 'root', $2, 1)", uuid.New().String(), hash)
	if err != nil {
		return nil, err
	}
	l.Debug("created root user")

	return db, nil
}
