package controller

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       uuid.UUID
	Username string
}

func CreateUser(ctx context.Context, g *app.Global, username, password string) (*User, error) {
	// Check if username is already in use
	exists, err := database.GetUserByName(ctx, g, username)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		// A user was found, so username is already in use
		return nil, nil
	}

	id := uuid.New()

	// bcrypt handles hash & salt automatically
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil
	}

	err = database.InsertUser(ctx, g, &database.User{
		ID:       id.String(),
		Username: username,
		Hash:     hash,
	})
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:       id,
		Username: username,
	}

	return user, nil
}
