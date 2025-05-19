package controller

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(ctx context.Context, g *app.Global,
	username, password string, isAdmin bool) (*models.User, error) {

	// Check if username is already in use
	exists, err := database.GetUserByName(ctx, g, username)
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

	err = database.InsertUser(ctx, g, user.ToDB())
	if err != nil {
		return nil, err
	}

	return user, nil
}
