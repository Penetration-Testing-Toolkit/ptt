package controller

import (
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func CreateProject(c echo.Context, g *app.Global, name string) (*model.Project, error) {
	// Check if name is already in use
	exists, err := database.GetProjectByName(c.Request().Context(), g, name)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		// A project was found, so name is already in use
		g.Logger().Debug("CreateProject: name already exists")
		return nil, nil
	}

	// Get user from session from echo.Context
	sess, err := GetSession(c)
	if err != nil {
		return nil, err
	}
	user := sess.User()

	project := model.NewProject(uuid.New(), name, user.ID, user.Username)

	err = database.InsertProject(c.Request().Context(), g, project.ToDB())
	if err != nil {
		return nil, err
	}

	return project, nil
}

func GetProjects(c echo.Context, g *app.Global) ([]*model.Project, error) {
	projectsDB, err := database.GetProjects(c.Request().Context(), g)
	if err != nil {
		return nil, err
	}

	// Convert models
	projects := make([]*model.Project, 0)
	for _, pDB := range projectsDB {
		p, err := model.ProjectFromDB(pDB)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func SelectProject(c echo.Context, g *app.Global, projectID string) (bool, error) {
	// Get session from echo.Context
	sess, err := GetSession(c)
	if err != nil {
		return false, err
	}

	// If projectID is empty, user wants to deselect project
	if len(projectID) == 0 {
		sess.SetProject(nil)
		return true, nil
	}

	// Parse project's ID (UUID)
	pID, err := uuid.Parse(projectID)
	if err != nil {
		return false, err
	}

	// Get project from db
	projectDB, err := database.GetProjectByID(c.Request().Context(), g, pID.String())
	if err != nil {
		return false, err
	}

	// Check for project
	if projectDB == nil {
		return false, nil
	}

	project, err := model.ProjectFromDB(projectDB)

	sess.SetProject(project)
	return true, nil
}

func DeleteProject(c echo.Context, g *app.Global, projectID string) (bool, error) {
	// Parse project's ID (UUID)
	pID, err := uuid.Parse(projectID)
	if err != nil {
		return false, err
	}

	// Get project from db
	project, err := database.GetProjectByID(c.Request().Context(), g, pID.String())
	if err != nil {
		return false, err
	}

	// Check for project
	if project == nil {
		return false, nil
	}

	// Get user from session from echo.Context
	sess, err := GetSession(c)
	if err != nil {
		return false, err
	}
	user := sess.User()

	// Project can only be deleted by owner or admin
	if project.OwnerID != user.ID.String() && !user.IsAdmin {
		return false, nil
	}

	// Delete project
	err = database.DeleteProject(c.Request().Context(), g, pID.String())
	if err != nil {
		return false, err
	}

	return true, nil
}
