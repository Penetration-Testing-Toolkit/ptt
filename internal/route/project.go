package route

import (
	"context"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template"
	"github.com/labstack/echo/v4"
	"net/http"
)

const minProjectNameLength = 3
const maxProjectNameLength = 30

// GetProject "GET /project".
func GetProject(c echo.Context, g *app.Global) Response {
	// Get list of projects from controller
	projects, err := controller.GetProjects(c, g)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Convert models
	projectsTempl := projectsToTempl(projects)

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Get user from session from echo.Context
	sess, err := controller.GetSession(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}
	userTempl := userTemplFromSession(sess)

	return Response{
		Component: Layout(c, g, template.GetProject(csrf, userTempl, projectsTempl)),
	}
}

// PostProjectCreate "POST /project/create".
func PostProjectCreate(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(projectCreateForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	project, err := controller.CreateProject(c, g, form.Name)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// No project created but there was no error -> name already exists
	if project == nil {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  template.Error("project name already exists"),
		}
	}

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Get user from session from echo.Context
	sess, err := controller.GetSession(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}
	userTempl := userTemplFromSession(sess)

	// Get list of projects from controller
	projects, err := controller.GetProjects(c, g)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Convert models
	projectsTempl := projectsToTempl(projects)

	// Success creating new project
	return Response{
		Component: template.CreateProjectSuccess(csrf, userTempl, projectsTempl),
	}
}

type projectCreateForm struct {
	Name string `form:"name"`
}

func (f *projectCreateForm) validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	if len(f.Name) < minProjectNameLength || maxProjectNameLength < len(f.Name) {
		problems = append(problems, fmt.Sprintf("project name must be between %d-%d characters long",
			minProjectNameLength, maxProjectNameLength))
	}

	return problems
}

// GetProjects "GET /project/projects"
func GetProjects(c echo.Context, g *app.Global) Response {
	// Get list of projects from controller
	projects, err := controller.GetProjects(c, g)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Convert models
	projectsTempl := projectsToTempl(projects)

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Get user from session from echo.Context
	sess, err := controller.GetSession(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}
	userTempl := userTemplFromSession(sess)

	return Response{
		Component: template.GetProjects(csrf, userTempl, projectsTempl),
	}
}

// PostProjectSelect "POST /project/select".
func PostProjectSelect(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(projectSelectForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	success, err := controller.SelectProject(c, g, form.ProjectID)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	if !success {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  template.Error("no project was selected"),
		}
	}

	return Response{
		StatusCode: http.StatusFound,
		Redirect:   "/",
	}
}

type projectSelectForm struct {
	ProjectID string `form:"projectID"`
}

func (f *projectSelectForm) validate(_ context.Context) (problems []string) {
	return []string{}
}

// PostProjectDelete "POST /project/delete".
func PostProjectDelete(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(deleteProjectForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	success, err := controller.DeleteProject(c, g, form.ProjectID)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	if !success {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  template.Error("no project was deleted"),
		}
	}

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Get user from session from echo.Context
	sess, err := controller.GetSession(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}
	userTempl := userTemplFromSession(sess)

	// Get list of projects from controller
	projects, err := controller.GetProjects(c, g)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Convert models
	projectsTempl := projectsToTempl(projects)

	return Response{
		Component: template.DeleteProjectSuccess(csrf, userTempl, projectsTempl),
	}
}

type deleteProjectForm struct {
	ProjectID string `form:"projectID"`
}

func (f *deleteProjectForm) validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	if len(f.ProjectID) == 0 {
		problems = append(problems, "projectID cannot be empty")
	}

	return problems
}
