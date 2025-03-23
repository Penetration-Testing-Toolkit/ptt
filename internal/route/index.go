package route

import (
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/templates"
	"github.com/labstack/echo/v4"
)

// GetIndex "GET /" returns the whole root page.
func GetIndex(_ echo.Context, g *app.Global) Response {
	i := plugin.ModulesToTemplate(g.Modules())

	return Response{
		Component: templates.Layout(i, templates.GetIndex(i)),
	}
}
