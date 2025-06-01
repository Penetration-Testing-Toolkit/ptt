package controller

import (
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func Login(c echo.Context, g *app.Global, username, password string) (*model.User, error) {
	// Find desired user in database
	userDB, err := database.GetUserByName(c.Request().Context(), g, username)
	if err != nil {
		return nil, err
	}

	if userDB == nil {
		// User does not exist
		g.Logger().Debug("Login: user does not exist")
		return nil, nil
	}

	// Convert model
	user, err := model.UserFromDB(userDB)
	if err != nil {
		return nil, err
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword(user.Hash, []byte(password))
	if err != nil {
		g.Logger().Debug("Login: password does not match")
		return nil, nil
	}

	// At this point, user has been authenticated

	// Delete previous session from session.Manager & invalidate cookie
	prevSession, err := GetSession(c)
	if err == nil {
		g.Sessions().DeleteSession(c, prevSession.ID())
	} else {
		g.Logger().Debug("Login: no previous session found", "error", err.Error())
	}

	s := g.Sessions().NewSession(c.Response(), user)

	g.Logger().Info("Login: user signed in", "userID", s.User().ID.String(), "username", s.User().Username)
	return user, nil
}
