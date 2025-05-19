package route

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/templates"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"unicode"
)

// GetAdmin "GET /admin/"
func GetAdmin(c echo.Context, g *app.Global) Response {
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	return Response{
		Component: Layout(c, g, templates.GetAdmin(csrf)),
	}
}

// PostAdminCreateUser "POST /admin/create-user"
func PostAdminCreateUser(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(createUserForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	user, err := controller.CreateUser(c.Request().Context(), g, form.Username, form.Password, form.IsAdmin)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// No user was created but there was no error -> username already exists
	if user == nil {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  templates.Error("username already exists"),
		}
	}

	// Success creating new user
	return Response{
		Component: templates.AdminCreateUserSuccess(),
	}
}

type createUserForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
	IsAdmin  bool   `form:"isAdmin"`
}

func (f *createUserForm) validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	// Treat usernames as all lowercase
	f.Username = strings.ToLower(f.Username)

	alphanumeric := true
	for _, c := range f.Username {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			alphanumeric = false
			break
		}
	}
	if !alphanumeric {
		problems = append(problems, "username must only contain alphanumeric characters")
	}

	if len(f.Username) < 3 {
		problems = append(problems, "username must be at least 3 characters long")
	}

	if len(f.Password) < 8 {
		problems = append(problems, "password must be at least 8 characters long")
	}

	return problems
}
