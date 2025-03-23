package plugin

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/templates"
	"github.com/Penetration-Testing-Toolkit/ptt/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// ModulePlugin contains everything the server needs to use a shared.Module plugin.
type ModulePlugin struct {
	// info is the shared.ModuleInfo about a plugin.
	info *shared.ModuleInfo

	// module is the shared.Module interface for the server to interact with.
	module shared.Module

	// client is the hashicorp/go-plugin plugin.Client.
	// These clients need to be terminated, which is done by the cleanupPlugins function.
	client *plugin.Client

	// rootURL is "/plugin/" + shared.ModuleInfo.ID.
	// A trailing '/' is not included as that will be included in the routes.
	rootURL *url.URL
}

var logger hclog.Logger

// StartPlugins begins the process of plugin discovery and registering routes
// Returns a slice of ModulePlugin, and a cleanup function that should be a called by defer.
// This way, the clients are cleaned when the server stops, rather than when this function returns.
func StartPlugins(hclog hclog.Logger, e *echo.Echo) (plugins []*ModulePlugin, cleanupFunc func([]*ModulePlugin)) {
	logger = hclog

	plugins = make([]*ModulePlugin, 0)

	// Get potential plugin executable paths
	files := discoverPlugins()

	// Attempt to start these plugin executables & get their info
	for _, file := range files {
		p, err := startPlugin(file)
		if err != nil {
			logger.Error("failed to start potential plugin", "filename", file, "err", err.Error())
		} else {
			plugins = append(plugins, p)
		}
	}

	// Register the shared.Module plugin's routes
	for _, plug := range plugins {
		err := registerRoutes(e, plug, plugins)
		if err != nil {
			logger.Error("error registering routes for plugin", "plugin_id", plug.info.ID, "err", err.Error())
		}
	}

	return plugins, cleanupPlugins
}

// discoverPlugins searches through the "plugins" directory for potential plugin executables to load.
// Returns a slice of file names that might be plugins.
func discoverPlugins() []string {
	var files = make([]string, 0)

	// Create "plugins" directory if it does not exist
	err := os.Mkdir("plugins", os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			logger.Info("'plugins' directory already exists")
		} else {
			logger.Error("could not create 'plugins' directory", "err", err.Error())
		}
	} else {
		logger.Info("'plugins' directory created")
	}

	// Read through "plugins" directory, look for files containing ".plugin" & add it as a potential plugin
	dir, err := os.ReadDir("plugins")
	if err != nil {
		logger.Error("could not read 'plugins' directory", "err", err.Error())
	}
	for _, f := range dir {
		if !f.IsDir() && strings.Contains(f.Name(), ".plugin") {
			logger.Debug("potential plugin found", "name", f.Name())
			files = append(files, f.Name())
		}
	}

	return files
}

// startPlugin starts plugin provided by filename.
func startPlugin(filename string) (*ModulePlugin, error) {
	// We're a host, start by launching the plugin process
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.HandshakeConfig,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("./plugins/" + filename),
		Logger:           logger.Named("plugins"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	// Connect via gRPC
	grpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, err
	}

	// Request the shared.Module interface from the plugin
	raw, err := grpcClient.Dispense("module")
	if err != nil {
		client.Kill()
		return nil, err
	}

	module := raw.(shared.Module)
	// Now we have a real shared.Module interface to work with

	// Get plugin's shared.ModuleInfo
	info, err := module.Register(context.Background())
	if err != nil {
		return nil, err
	}

	// All plugins are under the "/plugin/" path
	rootURL, err := url.Parse("/plugin/" + info.ID)
	if err != nil {
		return nil, err
	}

	return &ModulePlugin{
		info:    info,
		module:  module,
		client:  client,
		rootURL: rootURL,
	}, nil
}

// registerRoutes registers a plugin's shared.Module routes as an echo.Group.
// It registers these routes based on shared.ModuleInfo's routes.
func registerRoutes(e *echo.Echo, plugin *ModulePlugin, plugins []*ModulePlugin) error {
	info := plugin.info
	// Create group based on the plugin's ID
	g := e.Group(plugin.rootURL.String())

	for _, r := range info.Routes {
		// Setup echo.HandlerFunc
		handler := func(c echo.Context) error {
			if !r.UseSSE {
				// Regular (non-SSE) HTTP request
				// Proxy request to plugin's handler
				resp, err := plugin.module.Handle(c.Request().Context(), c.Request())
				if err != nil {
					return err
				}

				// Replace the existing response status & headers with plugin's response
				registerHelper(c, resp)

				return templates.Layout(ModulesToTemplate(plugins), templates.PluginContent(resp.Body)).Render(c.Request().Context(), c.Response())
			} else {
				// Handle SSE request (https://echo.labstack.com/docs/cookbook/sse)
				logger.Info("frontend SSE client connected", "IP", c.RealIP(), "Path", c.Request().URL.Path)

				// Setup headers for SSE
				h := c.Response().Header()
				h.Set("Content-Type", "text/event-stream")
				h.Set("Cache-Control", "no-cache")
				h.Set("Connection", "keep-alive")

				// Proxy request to plugin's SSE handler
				ch, err := plugin.module.HandleSSE(c.Request().Context(), c.Request())
				if err != nil {
					return err
				}

				for {
					select {
					case <-c.Request().Context().Done():
						logger.Info("frontend SSE client disconnected", "cause", "c.Request().Context().Done()", "IP", c.RealIP(), "Path", c.Request().URL.Path)
						return nil
					case resp, ok := <-ch:
						if !ok {
							logger.Debug("SSE channel was closed")
							return nil
						}

						logger.Debug("SSE response came through server's channel", "status", resp.Status, "resp", resp.Body)

						// Replace the existing response status & headers with plugin's response
						registerHelper(c, resp)

						// Convert response's body into an SSE & write to echo.Context response
						event := sseEvent{Data: []byte(resp.Body)}
						if err := event.marshalTo(c.Response()); err != nil {
							return err
						}

						// Sent response up to the frontend client
						c.Response().Flush()
						logger.Debug("SSE response flushed")
					}
				}
			}
		}

		// Dynamically map route's method to the correct echo route type
		methods := map[string]func(string, echo.HandlerFunc, ...echo.MiddlewareFunc) *echo.Route{
			http.MethodGet:     g.GET,
			http.MethodHead:    g.HEAD,
			http.MethodPost:    g.POST,
			http.MethodPut:     g.PUT,
			http.MethodDelete:  g.DELETE,
			http.MethodConnect: g.CONNECT,
			http.MethodOptions: g.OPTIONS,
			http.MethodTrace:   g.TRACE,
			http.MethodPatch:   g.PATCH,
		}
		if echoMethod, exists := methods[r.Method]; exists {
			echoMethod(r.Path, handler)
		} else {
			logger.Error("invalid HTTP method", "method", r.Method)
		}
	}

	return nil
}

// registerHelper overwrites a shared.Response status & headers into the echo.Context response.
func registerHelper(c echo.Context, resp *shared.Response) {
	c.Response().Status = resp.Status
	for k, v := range resp.Header {
		c.Response().Header().Del(k)
		for _, i := range v {
			c.Response().Header().Add(k, i)
		}
	}
}

// cleanupPlugins is a deferred function to kill & cleanup plugin clients.
func cleanupPlugins(plugins []*ModulePlugin) {
	logger.Debug("started killing plugin clients")
	for _, p := range plugins {
		logger.Debug("killing client", "plugin_id", p.info.ID)
		p.client.Kill()
	}
	logger.Debug("done killing plugin clients")
}

// ModulesToTemplate converts plugins slice into a format for the templ templates.
func ModulesToTemplate(modules []*ModulePlugin) []*templates.ModuleInfo {
	i := make([]*templates.ModuleInfo, 0)

	for _, plug := range modules {
		i = append(i, &templates.ModuleInfo{
			URL:     plug.rootURL.String(),
			Name:    plug.info.Name,
			Version: plug.info.Version,
		})
	}

	return i
}
