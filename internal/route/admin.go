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

// GetAdmin "GET /admin/".
func GetAdmin(c echo.Context, g *app.Global) Response {
	// Get initial users from controller
	users, err := controller.GetUsers(c, g, 10, 0)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	usersT := usersToTemplateUsers(users)

	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	return Response{
		Component: Layout(c, g, templates.GetAdmin(csrf, usersT, 10, 0)),
	}
}

// PostCreateUser "POST /admin/create-user".
func PostCreateUser(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(createUserForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	user, err := controller.CreateUser(c, g, form.Username, form.Password, form.IsAdmin)
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

// PostChangePassword "POST /admin/change-password".
func PostChangePassword(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(changePasswordForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	success, err := controller.ChangePassword(c, g, form.Username, form.OldPassword, form.NewPassword)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	if !success {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  templates.Error("password was not updated"),
		}
	}

	return Response{
		Component: templates.AdminChangePasswordSuccess(),
	}
}

type changePasswordForm struct {
	Username        string `form:"username"`
	OldPassword     string `form:"oldPassword"`
	NewPassword     string `form:"newPassword"`
	ConfirmPassword string `form:"confirmPassword"`
}

func (f *changePasswordForm) validate(_ context.Context) (problems []string) {
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

	if f.NewPassword != f.ConfirmPassword {
		problems = append(problems, "new password does not match confirm password")
	}

	if len(f.OldPassword) < 8 {
		problems = append(problems, "old password must be at least 8 characters long")
	}

	if len(f.NewPassword) < 8 {
		problems = append(problems, "new password must be at least 8 characters long")
	}

	return problems
}

// GetUsers "GET /admin/users".
func GetUsers(c echo.Context, g *app.Global) Response {
	// Parse query
	query := new(GetUsersQuery)
	resp := parse(c, g, query)
	if resp != nil {
		return *resp
	}

	// Send to controller
	users, err := controller.GetUsers(c, g, query.PageSize, query.Page)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	usersT := usersToTemplateUsers(users)

	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Success creating new user
	return Response{
		Component: templates.GetUsers(csrf, usersT, query.PageSize, query.Page),
	}
}

type GetUsersQuery struct {
	PageSize int `query:"pageSize"`
	Page     int `query:"page"`
}

func (g GetUsersQuery) validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	if g.PageSize < 10 || 50 < g.PageSize {
		problems = append(problems, "page size must be between [10. 50]")
	}

	if g.Page < 0 {
		problems = append(problems, "page must be > 0")
	}

	return problems
}

// PostDeleteUser "POST /admin/delete-user".
func PostDeleteUser(c echo.Context, g *app.Global) Response {
	// Parse form
	form := new(deleteUserForm)
	resp := parse(c, g, form)
	if resp != nil {
		return *resp
	}

	// Send to controller
	success, err := controller.DeleteUser(c, g, form.UserID)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	if !success {
		return Response{
			StatusCode: http.StatusUnprocessableEntity,
			Component:  templates.Error("no user was deleted"),
		}
	}

	return Response{
		Component: templates.AdminDeleteUserSuccess(),
	}
}

type deleteUserForm struct {
	UserID string `form:"userID"`
}

func (f *deleteUserForm) validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	if len(f.UserID) == 0 {
		problems = append(problems, "userID cannot be empty")
	}

	return problems
}
