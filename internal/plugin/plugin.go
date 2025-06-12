package plugin

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/model"
	"github.com/Penetration-Testing-Toolkit/ptt/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
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

func (m ModulePlugin) Info() *shared.ModuleInfo {
	return m.info
}

func (m ModulePlugin) Module() shared.Module {
	return m.module
}

func (m ModulePlugin) RootURL() *url.URL {
	return m.rootURL
}

// ModuleToTempl converts a ModulePlugin into a model.ModuleTempl.
func ModuleToTempl(module *ModulePlugin) *model.ModuleTempl {
	return &model.ModuleTempl{
		ID:       module.info.ID,
		URL:      module.rootURL.String(),
		Name:     module.info.Name,
		Version:  module.info.Version,
		Category: int(module.info.Category),
		Metadata: module.info.Metadata,
	}
}

// ModulesToTempl converts a ModulePlugin slice into a model.ModuleTempl slice.
func ModulesToTempl(modules []*ModulePlugin) []*model.ModuleTempl {
	i := make([]*model.ModuleTempl, 0)

	for _, plug := range modules {
		i = append(i, ModuleToTempl(plug))
	}

	return i
}

// StartPlugins begins the process of plugin discovery and starting plugins.
// Returns a slice of ModulePlugin, and a cleanup function that should be a called by defer.
// This way, the clients are cleaned when the server stops, rather than when this function returns.
func StartPlugins(logger hclog.Logger, storeServerAddr string) (plugins []*ModulePlugin,
	cleanupFunc func(hclog.Logger, []*ModulePlugin)) {

	plugins = make([]*ModulePlugin, 0)

	// Get potential plugin executable paths
	files := discoverPlugins(logger)

	// Attempt to start these plugin executables & get their info
	for _, file := range files {
		p, err := startPlugin(logger, file, storeServerAddr)
		if err != nil {
			logger.Error("failed to start potential plugin", "filename", file, "error", err.Error())
		} else {
			plugins = append(plugins, p)
		}
	}

	return plugins, cleanupPlugins
}

// discoverPlugins searches through the "plugins" directory for potential plugin executables to load.
// Returns a slice of file names that might be plugins.
func discoverPlugins(logger hclog.Logger) []string {
	files := make([]string, 0)

	// Create "plugins" directory if it does not exist
	err := os.Mkdir("plugins", os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			logger.Info("'plugins' directory already exists")
		} else {
			logger.Error("could not create 'plugins' directory", "error", err.Error())
		}
	} else {
		logger.Info("'plugins' directory created")
	}

	// Read through "plugins" directory, look for files containing ".plugin" & add it as a potential plugin
	dir, err := os.ReadDir("plugins")
	if err != nil {
		logger.Error("could not read 'plugins' directory", "error", err.Error())
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
func startPlugin(logger hclog.Logger, filename string, storeServerAddr string) (*ModulePlugin, error) {
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
	info, err := module.Register(context.Background(), storeServerAddr)
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

// cleanupPlugins is a deferred function to kill & cleanup plugin clients.
func cleanupPlugins(logger hclog.Logger, plugins []*ModulePlugin) {
	logger.Debug("started killing plugin clients")
	for _, p := range plugins {
		logger.Debug("killing client", "plugin_id", p.info.ID)
		p.client.Kill()
	}
	logger.Debug("done killing plugin clients")
}
