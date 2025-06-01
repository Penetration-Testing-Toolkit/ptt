package server

import (
	"errors"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/controller"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/route"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

// setupRoutes registers the app's main routes. This does not include plugin routes.
func setupRoutes(e *echo.Echo, g *app.Global) {
	// Adapts our handler style (route.HandlerFunc) into an echo.HandlerFunc
	// This allows us to pass the app.Global alongside echo.Context with a custom route.Response returned
	adapter := func(f route.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			resp := f(c, g)

			// Let the custom error handler deal with errors
			if resp.Err != nil {
				return resp.Err
			}

			// Handle redirects
			if resp.Redirect != "" {
				// If we ignored the status code, assume 302
				if resp.StatusCode == 0 {
					resp.StatusCode = http.StatusFound
				}
				err := c.Redirect(resp.StatusCode, resp.Redirect)
				if err != nil {
					return err
				}
				return nil
			}

			// At this point, we have a standard response
			// If we ignored the status code, assume 200
			if resp.StatusCode == 0 {
				resp.StatusCode = http.StatusOK
			}
			c.Response().Status = resp.StatusCode

			if resp.Component == nil {
				return errors.New("no response templ component given for request")
			}

			err := resp.Component.Render(c.Request().Context(), c.Response())
			if err != nil {
				return fmt.Errorf("error rendering templ component: %w", err)
			}

			return nil
		}
	}

	// CSRF protection middleware
	csrfMiddleware := middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLength:    64,
		TokenLookup:    "form:_csrf",
		CookieMaxAge:   300, // 5 min: CSRF tokens should be short-lived so they regenerate after inactivity
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		ErrorHandler: func(err error, c echo.Context) error { // CSRF has a separate error handler
			g.Logger().Error("CSRF error handler", "error", err.Error())
			c.Response().Status = http.StatusForbidden
			return route.Layout(c, g, template.ErrorPage(http.StatusForbidden, "Forbidden", "")).Render(c.Request().Context(), c.Response())
		},
	})

	// Index route
	e.GET("/", adapter(route.GetIndex))

	// "/login" & "/sign-out" sign-out routes
	e.GET("/login", adapter(route.GetLogin))
	e.POST("/login", adapter(route.PostLogin))
	e.GET("/sign-out", adapter(route.GetSignOut))

	// "/project" routes
	project := e.Group("/project")

	project.Use(csrfMiddleware)

	project.GET("", adapter(route.GetProject))
	project.GET("/projects", adapter(route.GetProjects))
	project.POST("/create", adapter(route.PostProjectCreate))
	project.POST("/select", adapter(route.PostProjectSelect))
	project.POST("/delete", adapter(route.PostProjectDelete))

	// "/admin" routes
	admin := e.Group("/admin", func(next echo.HandlerFunc) echo.HandlerFunc {
		// Admin authorization middleware
		return func(c echo.Context) error {
			sess, err := controller.GetSession(c)
			if err != nil {
				return err
			}

			if sess.User().IsAdmin {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden)
		}
	})

	admin.Use(csrfMiddleware)

	admin.GET("", adapter(route.GetAdmin))
	admin.GET("/users", adapter(route.GetUsers))
	admin.POST("/create-user", adapter(route.PostCreateUser))
	admin.POST("/change-password", adapter(route.PostChangePassword))
	admin.POST("/delete-user", adapter(route.PostDeleteUser))

	// Dev mode has some additional debug routes
	if g.DevMode() {
		debug := e.Group("/debug")
		// Route to simulate a 5xx internal server error
		debug.GET("/500", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusInternalServerError, "debug 500 error")
		})
	}
}
