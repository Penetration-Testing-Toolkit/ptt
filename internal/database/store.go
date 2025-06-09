package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
)

var storeInsert = `
INSERT INTO
	store (plugin_id, user_id, project_id, key, value)
VALUES
	($1, $2, $3, $4, $5)
;`

func StoreInsert(ctx context.Context, g *app.Global, pluginID, userID, projectID,
	key string, value []byte) error {

	var u *string = nil
	if len(userID) > 0 {
		u = &userID
	}

	var p *string = nil
	if len(projectID) > 0 {
		p = &projectID
	}

	result, err := g.DB().ExecContext(ctx, storeInsert, pluginID, u, p, key, value)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("StoreInsert: completed", "rowsAffected", rows)

	return nil
}

var storeUpdate = `
UPDATE
	store
SET
	value = $1
WHERE
	plugin_id == $2 AND user_id IS $3 AND project_id IS $4 AND key == $5
;`

func StoreUpdate(ctx context.Context, g *app.Global, pluginID, userID, projectID,
	key string, value []byte) error {

	var u *string = nil
	if len(userID) > 0 {
		u = &userID
	}

	var p *string = nil
	if len(projectID) > 0 {
		p = &projectID
	}

	result, err := g.DB().ExecContext(ctx, storeUpdate, value, pluginID, u, p, key)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("StoreUpdate: completed", "rowsAffected", rows)

	return nil
}

var storeGet = `
SELECT
	value
FROM
	store
WHERE
	plugin_id == $1 AND user_id IS $2 AND project_id IS $3 AND key == $4
LIMIT
	1
;`

func StoreGet(ctx context.Context, g *app.Global, pluginID, userID, projectID,
	key string) ([]byte, error) {

	var u *string = nil
	if len(userID) > 0 {
		u = &userID
	}

	var p *string = nil
	if len(projectID) > 0 {
		p = &projectID
	}

	value := make([]byte, 0)
	err := g.DB().GetContext(ctx, &value, storeGet, pluginID, u, p, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return value, nil
}

var storeDelete = `
DELETE FROM
	store
WHERE
	plugin_id == $1 AND user_id IS $2 AND project_id IS $3 AND key == $4
;`

func StoreDelete(ctx context.Context, g *app.Global, pluginID, userID, projectID,
	key string) error {

	var u *string = nil
	if len(userID) > 0 {
		u = &userID
	}

	var p *string = nil
	if len(projectID) > 0 {
		p = &projectID
	}

	result, err := g.DB().ExecContext(ctx, storeDelete, pluginID, u, p, key)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("StoreDelete: delete completed", "rowsAffected", rows)

	return nil
}
