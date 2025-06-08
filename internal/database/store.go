package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
)

var storeSet = `
INSERT OR REPLACE INTO
	store (plugin_id, key, value)
VALUES
	($1, $2, $3)
;`

func StoreSet(ctx context.Context, g *app.Global, pluginID, key string, value []byte) error {
	result, err := g.DB().ExecContext(ctx, storeSet, pluginID, key, value)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("StoreSet completed", "rowsAffected", rows)

	return nil
}

var storeGet = `
SELECT
	value
FROM
	store
WHERE
	plugin_id == $1 AND key == $2
LIMIT
	1
;`

func StoreGet(ctx context.Context, g *app.Global, pluginID, key string) ([]byte, error) {
	value := make([]byte, 0)
	err := g.DB().GetContext(ctx, &value, storeGet, pluginID, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return value, nil
}
