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

// GetLogin "GET /login".
func GetLogin(c echo.Context, g *app.Global) Response {
	return Response{
		Component: Layout(c, g, templates.GetLogin()),
	}
}

// PostLogin "POST /login".
func PostLogin(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(LoginForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	failed := "login failed"
	user, err := controller.Login(c, g, form.Username, form.Password)
	if err != nil {
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

	return Response{
		StatusCode: http.StatusFound,
		Redirect:   "/",
	}
}

type LoginForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (f *LoginForm) validate(_ context.Context) (problems []string) {
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

// GetSignOut "GET /sign-out".
func GetSignOut(c echo.Context, g *app.Global) Response {
	// Delete previous session
	s, err := controller.GetSession(c)
	if err == nil {
		g.Logger().Info("signing out", "userID", s.UserID(), "username", s.Username(),
			"username", s.Username())
		g.Sessions().DeleteSession(c, s.ID())
	} else {
		g.Logger().Debug("GetSignOut: error getting session", "error", err.Error())
	}

	// Redirect to "/login" page
	return Response{
		StatusCode: http.StatusFound,
		Redirect:   "/login",
	}
}
