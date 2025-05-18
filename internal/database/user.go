package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
)

type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	Hash     []byte `db:"hash"`
	IsAdmin  bool   `db:"is_admin"`
}

var insertUser = `INSERT INTO users (id, username, hash, is_admin) VALUES ($1, $2, $3, 0)`

func InsertUser(ctx context.Context, g *app.Global, user *User) error {
	result, err := g.DB().ExecContext(ctx, insertUser, user.ID, user.Username, user.Hash)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("insert user completed", "rows affected", rows)

	return nil
}

var getUserByID = `SELECT * from users WHERE id=$1 LIMIT 1`

func GetUserByID(ctx context.Context, g *app.Global, id string) (*User, error) {
	user := new(User)
	err := g.DB().GetContext(ctx, user, getUserByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows found, user ID does not exist
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func CheckUserIsAdmin(ctx context.Context, g *app.Global, userID string) (bool, error) {
	user := new(User)
	err := g.DB().GetContext(ctx, user, getUserByID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User ID not found
			return false, fmt.Errorf("CheckUserIsAdmin: user ID not found: %s", userID)
		}
		return false, err
	}

	if user.IsAdmin {
		return true, nil
	}

	return false, nil
}

var getUserByName = `SELECT * from users WHERE username=$1 LIMIT 1`

func GetUserByName(ctx context.Context, g *app.Global, username string) (*User, error) {
	user := new(User)
	err := g.DB().GetContext(ctx, user, getUserByName, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows found, username does not exist
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
