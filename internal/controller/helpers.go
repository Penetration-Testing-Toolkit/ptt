package controller

import (
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/session"
	"github.com/labstack/echo/v4"
)

// GetSession extracts the session.Session from the given echo.Context.
// The session.Manager middleware places the session.Session into the echo.Context if it exists.
func GetSession(c echo.Context) (*session.Session, error) {
	value := c.Get("session")
	if value == nil {
		return nil, errors.New("GetSession: session not found in echo context")
	}

	s, ok := value.(*session.Session)
	if !ok {
		return nil, errors.New("GetSession: echo context's session is not a session")
	}

	return s, nil
}
