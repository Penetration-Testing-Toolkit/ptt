package route

import (
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template"
	"github.com/labstack/echo/v4"
)

// GetIndex "GET /" returns the whole root page.
func GetIndex(c echo.Context, g *app.Global) Response {
	sess, err := controller.GetSession(c)
	if err != nil {
		return Response{
			Err: err,
		}
	}

	return Response{
		Component: Layout(c, g, template.GetIndex(plugin.ModulesToTempl(g.Modules()),
			userTemplFromSession(sess))),
	}
}
