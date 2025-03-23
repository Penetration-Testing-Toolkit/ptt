package app

import (
	"context"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/plugin"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/session"
	"github.com/hashicorp/go-hclog"
	"github.com/jmoiron/sqlx"
)

// Global contains globals used throughout the app.
// It is placed in its own package to prevent cyclic import errors.
type Global struct {
	logger   hclog.Logger
	db       *sqlx.DB
	sessions *session.Manager
	modules  []*plugin.ModulePlugin
	devMode  bool
}

func NewGlobal(logger hclog.Logger, db *sqlx.DB, sessions *session.Manager, modules []*plugin.ModulePlugin, devMode bool) Global {
	return Global{
		logger:   logger,
		db:       db,
		sessions: sessions,
		modules:  modules,
		devMode:  devMode,
	}
}

func (g *Global) Logger() hclog.Logger {
	return g.logger
}

func (g *Global) DB() *sqlx.DB {
	return g.db
}

func (g *Global) Sessions() *session.Manager {
	return g.sessions
}

func (g *Global) Modules() []*plugin.ModulePlugin {
	return g.modules
}

func (g *Global) DevMode() bool {
	return g.devMode
}

type Validator interface {
	Validate(ctx context.Context) (problems []string)
}
