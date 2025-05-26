package controller

import (
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c echo.Context, g *app.Global,
	username, password string, isAdmin bool) (*models.User, error) {

	// Check if username is already in use
	exists, err := database.GetUserByName(c.Request().Context(), g, username)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		// A user was found, so username is already in use
		g.Logger().Debug("CreateUser: username already exists")
		return nil, nil
	}

	// bcrypt handles hash & salt automatically
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.NewUser(uuid.New(), username, hash, isAdmin)

	err = database.InsertUser(c.Request().Context(), g, user.ToDB())
	if err != nil {
		return nil, err
	}

	return user, nil
}

func ChangePassword(c echo.Context, g *app.Global,
	username, oldPassword, newPassword string) (bool, error) {

	// Get user
	user, err := database.GetUserByName(c.Request().Context(), g, username)
	if err != nil {
		return false, err
	}
	if user == nil {
		g.Logger().Debug("ChangePassword: no user found")
		return false, nil
	}

	// Compare database's hash to provided old password
	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(oldPassword))
	if err != nil {
		g.Logger().Debug("ChangePassword: password does not match")
		return false, nil
	}

	// At this point, user has the correct old password

	// bcrypt handles hash & salt automatically
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	err = database.ChangePassword(c.Request().Context(), g, string(hash), user.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetUsers(c echo.Context, g *app.Global, pageSize, page int) ([]*models.User, error) {
	usersDB, err := database.GetUsers(c.Request().Context(), g, pageSize, page*pageSize)
	if err != nil {
		return nil, err
	}

	// Convert models
	users := make([]*models.User, 0)
	for _, uDB := range usersDB {
		u, err := models.UserFromDB(uDB)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func DeleteUser(c echo.Context, g *app.Global, id string) (bool, error) {
	// Parse user's ID (UUID)
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Get user from db
	user, err := database.GetUserByID(c.Request().Context(), g, uid.String())
	if err != nil {
		return false, err
	}

	// Check for user
	if user == nil {
		return false, nil
	}

	// Prevent deleting default "root" user
	if user.Username == "root" {
		return false, nil
	}

	// Delete user
	err = database.DeleteUser(c.Request().Context(), g, uid.String())
	if err != nil {
		return false, err
	}

	return true, nil
}
