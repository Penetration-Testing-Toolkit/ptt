package route

import (
	"context"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/session"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"net/http"
)

// HandlerFunc is a custom handler type to be used in this app.
// This allows for passing app.Global into our handlers.
type HandlerFunc func(echo.Context, *app.Global) Response

// Response is a custom HTTP response that uses templ.Components.
type Response struct {
	StatusCode int
	Component  templ.Component
	Redirect   string
	Err        error
}

type validatable interface {
	validate(ctx context.Context) (problems []string)
}

// Layout wraps the given templ.Component in the template.Layout page.
func Layout(c echo.Context, g *app.Global, component templ.Component) templ.Component {
	modulesTempl := plugin.ModulesToTempl(g.Modules())

	sess, err := controller.GetSession(c)
	if err != nil {
		// If no session found, treat as no account
		return template.Layout(modulesTempl, nil, nil, component)
	}

	return template.Layout(modulesTempl, userTemplFromSession(sess), projectTemplFromSession(sess), component)
}

// userTemplFromSession returns the model.UserTempl from the given session.Session.
func userTemplFromSession(s *session.Session) *model.UserTempl {
	return &model.UserTempl{
		ID:       s.User().ID.String(),
		Username: s.User().Username,
		IsAdmin:  s.User().IsAdmin,
	}
}

// projectTemplFromSession returns the model.ProjectTempl from the given session.Session.
func projectTemplFromSession(s *session.Session) *model.ProjectTempl {
	if s.Project() == nil {
		return nil
	}

	return &model.ProjectTempl{
		ID:        s.Project().ID.String(),
		Name:      s.Project().Name,
		OwnerID:   s.Project().OwnerID.String(),
		OwnerName: s.Project().OwnerName,
	}
}

// usersToTempl converts a slice of models.User into a slice of model.UserTempl.
func usersToTempl(users []*model.User) []*model.UserTempl {
	usersTempl := make([]*model.UserTempl, 0)
	for _, user := range users {
		usersTempl = append(usersTempl, user.ToTempl())
	}
	return usersTempl
}

func projectsToTempl(projects []*model.Project) []*model.ProjectTempl {
	projectsTempl := make([]*model.ProjectTempl, 0)
	for _, p := range projects {
		projectsTempl = append(projectsTempl, p.ToTempl())
	}
	return projectsTempl
}

// parse binds path params, query params, and the request body to a validatable struct and validates its input.
// Returns a Response if there was a problem, else it returns nil.
func parse(c echo.Context, g *app.Global, form validatable) *Response {
	// Bind form
	if err := c.Bind(form); err != nil {
		g.Logger().Debug("parse: cound not bind form", "error", err.Error())
		return &Response{
			StatusCode: http.StatusBadRequest,
			Component:  template.InvalidInput([]string{"could not process form"}),
		}
	}

	// Validate input
	problems := form.validate(c.Request().Context())
	if len(problems) != 0 {
		g.Logger().Debug("parse: invalid input")
		return &Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  template.InvalidInput(problems),
		}
	}

	return nil
}

// getCSRF extracts the CSRF string from the given echo.Context.
// This CSRF string derives from the echo CSRF middleware used in the "/admin/*" routes.
// This string should be passed inside HTML forms sent to the user.
func getCSRF(c echo.Context) (string, error) {
	value := c.Get("csrf")
	csrf, ok := value.(string)
	if !ok {
		return "", errors.New("getCSRF: echo context's csrf key was not a string")
	}

	return csrf, nil
}
