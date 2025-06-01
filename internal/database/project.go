package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
)

var insertProject = `
INSERT INTO
	projects (id, name, owner_id)
VALUES
	($1, $2, $3)
;`

func InsertProject(ctx context.Context, g *app.Global, project *model.ProjectDB) error {
	result, err := g.DB().ExecContext(ctx, insertProject, project.ID, project.Name, project.OwnerID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("insert project completed", "rowsAffected", rows)

	return nil
}

var getProjectByID = `
SELECT
	projects.id, name, owner_id, users.username as owner_name
FROM
	projects
INNER JOIN
	users ON users.id == projects.owner_id
WHERE
	projects.id == $1
LIMIT
	1
;`

func GetProjectByID(ctx context.Context, g *app.Global, id string) (*model.ProjectDB, error) {
	project := new(model.ProjectDB)
	err := g.DB().GetContext(ctx, project, getProjectByID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows found, project ID does not exist
			return nil, nil
		}
		return nil, err
	}

	return project, nil
}

var getProjectByName = `
SELECT
	projects.id, name, owner_id, users.username as owner_name
FROM
	projects
INNER JOIN
	users ON users.id == projects.owner_id
WHERE
	name == $1
LIMIT
	1
;`

func GetProjectByName(ctx context.Context, g *app.Global, name string) (*model.ProjectDB, error) {
	project := new(model.ProjectDB)
	err := g.DB().GetContext(ctx, project, getProjectByName, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows found, project name does not exist
			return nil, nil
		}
		return nil, err
	}

	return project, nil
}

var getProjects = `
SELECT
	projects.id, name, owner_id, users.username as owner_name
FROM
	projects
INNER JOIN
	users ON users.id == projects.owner_id
;`

func GetProjects(ctx context.Context, g *app.Global) ([]*model.ProjectDB, error) {
	projects := make([]*model.ProjectDB, 0)
	err := g.DB().SelectContext(ctx, &projects, getProjects)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No rows -> no projects
			return projects, nil
		}
		return nil, err
	}

	return projects, nil
}

var deleteProject = `
DELETE FROM
	projects
WHERE
	id == $1
;`

func DeleteProject(ctx context.Context, g *app.Global, id string) error {
	result, err := g.DB().ExecContext(ctx, deleteProject, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	g.Logger().Debug("delete project completed", "rowsAffected", rows)

	return nil
}
