package route

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/templates"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// GetLogin "GET /login"
func GetLogin(_ echo.Context, g *app.Global) Response {
	return Response{
		Component: Layout(g, templates.GetLogin()),
	}
}

type LoginForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (f *LoginForm) Validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	// Treat usernames as all lowercase
	f.Username = strings.ToLower(f.Username)

	if len(f.Username) == 0 {
		problems = append(problems, "username cannot be empty")
	}

	if len(f.Password) == 0 {
		problems = append(problems, "password cannot be empty")
	}

	return problems
}

// PostLogin "POST /login"
func PostLogin(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(LoginForm)
	if err := c.Bind(form); err != nil {
		g.Logger().Debug("PostLogin: could not bind form", "err", err.Error())
		return Response{
			StatusCode: http.StatusBadRequest,
			Component:  templates.InvalidInput([]string{"could not process form"}),
		}
	}

	// Validate input
	problems := form.Validate(c.Request().Context())
	if len(problems) != 0 {
		g.Logger().Debug("PostLogin: invalid credentials")
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  templates.InvalidInput(problems),
		}
	}

	// Send to controller
	failed := "login failed"
	user, err := controller.Login(c.Request().Context(), g, form.Username, form.Password)
	if err != nil {
		g.Logger().Error("PostLogin: error logging in user", "err", err.Error())
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  templates.Error(failed),
		}
	}

	// Wrong password or user does not exist
	if user == nil {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  templates.Error(failed),
		}
	}

	// Delete previous session
	session, err := GetSession(c)
	if err == nil {
		g.Sessions().DeleteSession(session.ID())
	}

	g.Sessions().NewSession(c.Response(), user.ID)

	g.Logger().Debug("login successful, redirecting")
	return Response{
		StatusCode: http.StatusFound,
		Redirect:   "/",
	}
}
