package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/app"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/route"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/session"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/template"
	"github.com/Penetration-Testing-Toolkit/ptt/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Config struct {
	Logger   hclog.Logger
	Static   fs.FS
	Address  string
	DB       *sqlx.DB
	Sessions *session.Manager
	DevMode  bool
}

func Start(cfg *Config) {
	l := cfg.Logger

	g := new(app.Global)
	*g = app.NewGlobal(l, cfg.DB, cfg.Sessions, make([]*plugin.ModulePlugin, 0), cfg.DevMode)

	// Create Echo server
	e := echo.New()

	// Hide the Echo banner
	e.HideBanner = true
	e.HidePort = true

	// Format echo's request logging through hashicorp/go-hclog
	requestLogger := l.Named("request")
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			// Log everything when at debug level
			if g.Logger().GetLevel() <= hclog.Debug {
				return false
			}

			// Else, skip logging static assets
			if strings.HasPrefix(c.Request().URL.Path, "/static") || c.Request().URL.Path == "/favicon.ico" {
				return true
			}

			return false
		},
		LogRemoteIP: true,
		LogURI:      true,
		LogMethod:   true,
		LogProtocol: true,
		LogStatus:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestLogger.With("IP", v.RemoteIP).
				With("URI", v.URI).
				With("method", v.Method).
				With("protocol", v.Protocol).
				With("status", v.Status).
				Info("new request")
			return nil
		},
	}))

	// Custom echo error handler using our hclog.Logger
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		l.Debug("error handler", "error", err.Error())
		if c.Response().Committed {
			return
		}

		code := http.StatusInternalServerError
		message := ""

		// Try and parse as an echo.HTTPError
		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
			if s, ok := he.Message.(string); ok {
				message = s
			}
		}
		c.Response().Status = code

		// Common error codes have their own page
		switch code {
		case http.StatusUnauthorized:
			err = route.Layout(c, g, template.ErrorPage(http.StatusUnauthorized, "Unauthorized", message)).Render(c.Request().Context(), c.Response())
		case http.StatusForbidden:
			err = route.Layout(c, g, template.ErrorPage(http.StatusForbidden, "Forbidden", message)).Render(c.Request().Context(), c.Response())
		case http.StatusNotFound:
			err = route.Layout(c, g, template.ErrorPage(http.StatusNotFound, "Not Found", message)).Render(c.Request().Context(), c.Response())
		case http.StatusInternalServerError:
			fallthrough
		default:
			err = route.Layout(c, g, template.ErrorPage(http.StatusInternalServerError, "Internal Server Error", message)).Render(c.Request().Context(), c.Response())
		}

		if err == nil {
			return
		} else {
			// Something went really wrong
			code = http.StatusInternalServerError
			l.Error("internal server error: could not render an error page", "code", code, "error", err.Error())
			c.Response().Status = code
			err = c.String(code, "something went wrong, internal server error")
			if err != nil {
				return
			}
		}
	}

	// Security middleware
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection: "",     // Deprecated, leave blank to keep disabled!
		XFrameOptions: "DENY", // Disable site from being displayed in iFrames
		// TODO: CSP header
	}))

	// Session middleware
	e.Use(cfg.Sessions.Middleware)

	// Host static assets
	staticServer := http.FileServer(http.FS(cfg.Static))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", staticServer)))
	e.GET("/favicon.ico", echo.WrapHandler(staticServer))

	// Setup server's routes
	setupRoutes(e, g)

	// Start plugin discovery
	shared.Logger = l
	plugins, cleanup := plugin.StartPlugins(l)
	defer cleanup(l, plugins)

	// Register plugin routes
	for _, plug := range plugins {
		err := route.RegisterPluginRoutes(l, e, g, plug)
		if err != nil {
			l.Error("error registering routes for plugin", "pluginID", plug.Info().ID, "error", err.Error())
		}
	}

	// Update global to include the plugins list
	*g = app.NewGlobal(l, cfg.DB, cfg.Sessions, plugins, cfg.DevMode)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
	go func() {
		if g.Logger().GetLevel() <= hclog.Debug {
			for _, r := range e.Routes() {
				l.Debug("route defined", "method", fmt.Sprintf("%v", r.Method),
					"path", fmt.Sprintf("%v", r.Path))
			}
		}

		l.Info("starting server", "address", cfg.Address)
		if err := e.Start(cfg.Address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("shutting down the server")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		l.Error("error shutting down server", "error", err.Error())
		os.Exit(1)
	}
}
