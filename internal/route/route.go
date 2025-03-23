package route

import (
	"errors"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/session"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/templates"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type HandlerFunc func(echo.Context, *app.Global) Response

type Response struct {
	StatusCode int
	Component  templ.Component
	Redirect   string
	Err        error
}

func Layout(g *app.Global, component templ.Component) templ.Component {
	i := plugin.ModulesToTemplate(g.Modules())

	return templates.Layout(i, component)
}

func GetSession(c echo.Context) (*session.Session, error) {
	value := c.Get("session")
	if value == nil {
		return nil, errors.New("session not found in echo context")
	}

	s, ok := value.(*session.Session)
	if !ok {
		return nil, errors.New("context session is not a session")
	}

	return s, nil
}

func GetCSRF(c echo.Context) (string, error) {
	value := c.Get("csrf")
	csrf, ok := value.(string)
	if !ok {
		return "", errors.New("echo context csrf was not a string")
	}

	return csrf, nil
}
