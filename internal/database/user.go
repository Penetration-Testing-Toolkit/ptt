package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/models"
)

var insertUser = `INSERT INTO users (id, username, hash, is_admin) VALUES ($1, $2, $3, $4)`

func InsertUser(ctx context.Context, g *app.Global, user *models.UserDB) error {
	result, err := g.DB().ExecContext(ctx, insertUser, user.ID, user.Username, user.Hash, user.IsAdmin)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("insert user completed", "rowsAffected", rows)

	return nil
}

var getUserByID = `SELECT * from users WHERE id=$1 LIMIT 1`

func GetUserByID(ctx context.Context, g *app.Global, id string) (*models.UserDB, error) {
	user := new(models.UserDB)
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

var getUserByName = `SELECT * from users WHERE username=$1 LIMIT 1`

func GetUserByName(ctx context.Context, g *app.Global, username string) (*models.UserDB, error) {
	user := new(models.UserDB)
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
