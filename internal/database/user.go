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

var transferProjectOwner = `
UPDATE
	projects
SET
	owner_id = $1
WHERE
	owner_id == $2
;`

var deleteUserFromStore = `
DELETE FROM
	store
WHERE
	project_id IS NULL AND user_id == $1
;`

var deleteUser = `
DELETE FROM
	users
WHERE
	id == $1
;`

func DeleteUser(ctx context.Context, g *app.Global, id string) error {
	// If user owned any projects, transfer ownership to "root" admin
	root, err := GetUserByName(ctx, g, "root")
	if err != nil {
		return err
	}

	result, err := g.DB().ExecContext(ctx, transferProjectOwner, root.ID, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("DeleteUser: transfer project owner completed", "rowsAffected", rows)

	// Delete user from store table where there is no project
	// (i.e. value is associated with the user but not a project)
	result, err = g.DB().ExecContext(ctx, deleteUserFromStore, id)
	if err != nil {
		return err
	}

	rows, err = result.RowsAffected()
	g.Logger().Debug("DeleteUser: delete user from store completed", "rowsAffected", rows)

	// Delete user from users table
	result, err = g.DB().ExecContext(ctx, deleteUser, id)
	if err != nil {
		return err
	}

	rows, err = result.RowsAffected()
	g.Logger().Debug("DeleteUser: delete user completed", "rowsAffected", rows)

	return nil
}
