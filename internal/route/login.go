package route

import (
	"context"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// GetLogin "GET /login".
func GetLogin(c echo.Context, g *app.Global) Response {
	return Response{
		Component: Layout(c, g, template.GetLogin(g.DevMode())),
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
			Component:  template.Error(failed),
		}
	}

	// Wrong password or user does not exist
	if user == nil {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  template.Error(failed),
		}
	}

	return Response{
		StatusCode: http.StatusFound,
		Redirect:   "/project",
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

	if len(f.Username) < minUsernameLength || maxUsernameLength < len(f.Username) {
		problems = append(problems, fmt.Sprintf("username must be between %d-%d characters long",
			minUsernameLength, maxUsernameLength))
	}

	if len(f.Password) < minPasswordLength || maxPasswordLength < len(f.Password) {
		problems = append(problems, fmt.Sprintf("password must be between %d-%d characters long",
			minPasswordLength, maxPasswordLength))
	}

	return problems
}

// GetSignOut "GET /sign-out".
func GetSignOut(c echo.Context, g *app.Global) Response {
	// Delete previous session
	prev, err := controller.GetSession(c)
	if err == nil {
		g.Logger().Info("signing out", "userID", prev.User().ID.String(), "username", prev.User().Username)
		g.Sessions().DeleteSession(c, prev.ID())
	} else {
		g.Logger().Debug("GetSignOut: error getting session", "error", err.Error())
	}

	// Redirect to "/login" page
	return Response{
		StatusCode: http.StatusFound,
		Redirect:   "/login",
	}
}
