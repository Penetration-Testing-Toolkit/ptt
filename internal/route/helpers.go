package route

import (
	"context"
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/session"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/templates"
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

// Layout wraps the given templ.Component in the templates.Layout page.
func Layout(c echo.Context, g *app.Global, component templ.Component) templ.Component {
	i := plugin.ModulesToTemplate(g.Modules())

	s, err := controller.GetSession(c)
	if err != nil {
		// If no session found, treat as no account
		return templates.Layout(i, nil, component)
	}

	return templates.Layout(i, sessionToTemplateUser(s), component)
}

// sessionToTemplateUser converts a session.Session into a templates.User struct.
func sessionToTemplateUser(s *session.Session) *templates.User {
	return &templates.User{
		ID:       s.UserID().String(),
		Username: s.Username(),
		IsAdmin:  s.IsAdmin(),
	}
}

// parse binds a POST form to a validatable struct and validates its input.
// Returns a Response if there was a problem, else it returns nil.
func parse(c echo.Context, g *app.Global, form validatable) *Response {
	// Bind form
	if err := c.Bind(form); err != nil {
		g.Logger().Error("parse: cound not bind form", "error", err.Error())
		return &Response{
			StatusCode: http.StatusBadRequest,
			Component:  templates.InvalidInput([]string{"could not process form"}),
		}
	}

	// Validate input
	problems := form.validate(c.Request().Context())
	if len(problems) != 0 {
		g.Logger().Error("parse: invalid input")
		return &Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  templates.InvalidInput(problems),
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
