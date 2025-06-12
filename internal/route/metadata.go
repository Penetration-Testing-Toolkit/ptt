package route

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template"
	"github.com/labstack/echo/v4"
)

// GetMetadata "GET /metadata".
func GetMetadata(c echo.Context, g *app.Global) Response {
	// Parse query
	query := new(GetMetadataQuery)
	resp := parse(c, g, query)
	if resp != nil {
		return *resp
	}

	var m *model.ModuleTempl = nil

	// Look for matching plugin
	for _, plug := range g.Modules() {
		if plug.Info().ID == query.ID {
			m = plugin.ModuleToTempl(plug)
			break
		}
	}

	if m == nil {
		return Response{
			Component: template.InvalidInput([]string{"could not find plugin with matching ID"}),
		}
	}

	return Response{
		Component: Layout(c, g, template.GetMetadata(m)),
	}
}

type GetMetadataQuery struct {
	ID string `query:"id"`
}

func (q *GetMetadataQuery) validate(_ context.Context) (problems []string) {
	problems = make([]string, 0)

	if len(q.ID) == 0 {
		problems = append(problems, "plugin ID cannot be empty")
	}

	return problems
}
