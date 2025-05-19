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

var schema = `CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY NOT NULL,
	username TEXT NOT NULL UNIQUE,
	hash BLOB NOT NULL,
	is_admin INT NOT NULL
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

	// Create default admin account
	hash, err := bcrypt.GenerateFromPassword([]byte("CHANGE_ME!!"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("INSERT INTO users VALUES ($1, 'admin', $2, 1)", uuid.New().String(), hash)
	if err != nil {
		return nil, err
	}
	l.Debug("created default admin account")

	return db, nil
}
