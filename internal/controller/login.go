package controller

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Login(ctx context.Context, g *app.Global, username, password string) (*User, error) {
	// Find desired user in database
	user, err := database.GetUserByName(ctx, g, username)
	if err != nil {
		return nil, err
	}

	// User does not exist
	if user == nil {
		return nil, nil
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(password))
	if err != nil {
		return nil, nil
	}

	id, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: username,
	}, nil
}
