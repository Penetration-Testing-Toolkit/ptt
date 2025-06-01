package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
)

var insertUser = `
INSERT INTO
	users (id, username, hash, is_admin)
VALUES
	($1, $2, $3, $4)
;`

func InsertUser(ctx context.Context, g *app.Global, user *model.UserDB) error {
	result, err := g.DB().ExecContext(ctx, insertUser, user.ID, user.Username, user.Hash, user.IsAdmin)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("insert user completed", "rowsAffected", rows)

	return nil
}

var getUserByID = `
SELECT
	id, username, hash, is_admin
FROM
	users
WHERE
	id == $1
LIMIT
	1
;`

func GetUserByID(ctx context.Context, g *app.Global, id string) (*model.UserDB, error) {
	user := new(model.UserDB)
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

var getUserByName = `
SELECT
	id, username, hash, is_admin
FROM
	users
WHERE
	username == $1
LIMIT
	1
;`

func GetUserByName(ctx context.Context, g *app.Global, username string) (*model.UserDB, error) {
	user := new(model.UserDB)
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

var getUsers = `
SELECT
	id, username, hash, is_admin
FROM
	users
LIMIT
	$1
OFFSET
	$2
;`

func GetUsers(ctx context.Context, g *app.Global, limit, offset int) ([]*model.UserDB, error) {
	users := make([]*model.UserDB, 0)
	err := g.DB().SelectContext(ctx, &users, getUsers, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No rows -> no users
			return users, nil
		}
		return nil, err
	}

	return users, nil
}

var changePassword = `
UPDATE
	users
SET
	hash = $1
WHERE
	id == $2
;`

func ChangePassword(ctx context.Context, g *app.Global, hash, id string) error {
	result, err := g.DB().ExecContext(ctx, changePassword, hash, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("change password completed", "rowsAffected", rows)

	return nil
}

var deleteUser = `
DELETE FROM
	users
WHERE
	id == $1
;`

func DeleteUser(ctx context.Context, g *app.Global, id string) error {
	result, err := g.DB().ExecContext(ctx, deleteUser, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("delete user completed", "rowsAffected", rows)

	return nil
}
