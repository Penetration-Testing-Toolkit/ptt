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
	"unicode"
)

const minUsernameLength = 3
const maxUsernameLength = 20
const minPasswordLength = 8
const maxPasswordLength = 72 // bcrypt password cannot exceed 72 bytes

// GetAdmin "GET /admin".
func GetAdmin(c echo.Context, g *app.Global) Response {
	// Get list of users from controller
	users, err := controller.GetUsers(c, g, 10, 0)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Convert models
	usersTempl := usersToTempl(users)

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	return Response{
		Component: Layout(c, g, template.GetAdmin(csrf, usersTempl, 10, 0)),
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
			Component:  template.Error("username already exists"),
		}
	}

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Get list of users from controller
	users, err := controller.GetUsers(c, g, 10, 0)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Convert models
	usersTempl := usersToTempl(users)

	// Success creating new user
	return Response{
		Component: template.CreateUserSuccess(csrf, usersTempl),
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

	if len(f.Username) < minUsernameLength || maxUsernameLength < len(f.Username) {
		problems = append(problems, fmt.Sprintf("username must be between %d-%d characters long",
			minUsernameLength, maxUsernameLength))
	} else {
		// If username is within acceptable length, check for invalid characters
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
	}

	if len(f.Password) < minPasswordLength || maxPasswordLength < len(f.Password) {
		problems = append(problems, fmt.Sprintf("password must be between %d-%d characters long",
			minPasswordLength, maxPasswordLength))
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
			Component:  template.Error("password was not updated"),
		}
	}

	return Response{
		Component: template.ChangePasswordSuccess(),
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

	if len(f.Username) < minUsernameLength || maxUsernameLength < len(f.Username) {
		problems = append(problems, fmt.Sprintf("username must be between %d-%d characters long",
			minUsernameLength, maxUsernameLength))
	} else {
		// If username is within acceptable length, check for invalid characters
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
	}

	if f.NewPassword != f.ConfirmPassword {
		problems = append(problems, "new password does not match confirm password")
	}

	if len(f.OldPassword) < minPasswordLength || maxPasswordLength < len(f.OldPassword) {
		problems = append(problems, fmt.Sprintf("old password must be between %d-%d characters long",
			minPasswordLength, maxPasswordLength))
	}

	if len(f.NewPassword) < minPasswordLength || maxPasswordLength < len(f.NewPassword) {
		problems = append(problems, fmt.Sprintf("new password must be between %d-%d characters long",
			minPasswordLength, maxPasswordLength))
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

	// Convert models
	usersTempl := usersToTempl(users)

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Success creating new user
	return Response{
		Component: template.GetUsers(csrf, usersTempl, query.PageSize, query.Page),
	}
}

type GetUsersQuery struct {
	PageSize int `query:"pageSize"`
	Page     int `query:"page"`
}

func (q *GetUsersQuery) validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	if q.PageSize < 10 || 50 < q.PageSize {
		problems = append(problems, "page size must be between [10. 50]")
	}

	if q.Page < 0 {
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
			Component:  template.Error("no user was deleted"),
		}
	}

	// Get CSRF from echo.Context
	csrf, err := getCSRF(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Get list of users from controller
	users, err := controller.GetUsers(c, g, 10, 0)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	// Convert models
	usersTempl := usersToTempl(users)

	return Response{
		Component: template.DeleteUserSuccess(csrf, usersTempl),
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
