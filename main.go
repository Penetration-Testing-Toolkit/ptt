package main

import (
	"embed"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/database"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/server"
	"github.com/Penetration-Testing-Toolkit/ptt/internal/session"
	"github.com/Penetration-Testing-Toolkit/ptt/shared"
	"github.com/hashicorp/go-hclog"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"
)

//go:embed static
var embeddedFS embed.FS

func main() {
	// Read environment variables to configure logging
	json := true
	if strings.ToUpper(os.Getenv("JSON")) == "FALSE" {
		json = false
	}
	logLevel := shared.LoggerOptions.Level
	levelStr := os.Getenv("LOG_LEVEL")
	level, err := strconv.Atoi(levelStr)
	if err == nil {
		logLevel = hclog.Level(level)
	}

	// Create a hclog.Logger
	// TODO: split log output to Stdout AND a log file. Bonus: hash the log file during server shutdown for integrity
	l := hclog.New(&hclog.LoggerOptions{
		Name:       "ptt",
		Level:      logLevel,
		Output:     os.Stdout,
		JSONFormat: json,
	})
	l.Debug("logger created", "level", logLevel, "json", json)

	// Determine if files will be hosted from embedded in the binary or from the OS file system
	var f fs.FS
	env := os.Getenv("ENV")
	devMode := strings.ToUpper(env) == "DEV"
	if devMode {
		// If dev environment, host static assets from the OS file system
		f = os.DirFS("static")
		l.Debug("hosting static files from file system")
	} else {
		// Else, host static assets embedded in Go binary
		f, err = fs.Sub(embeddedFS, "static")
		if err != nil {
			l.Error("error finding 'static' directory in embedded filesystem", "error", err.Error())
			panic(err)
		}
		l.Debug("hosting static files from embedded file system")
	}

	// Setup database
	db, err := database.SetupDB(l)
	if err != nil {
		l.Error("error setting up database", "error", err.Error())
		panic(err)
	}

	// Setup session manager
	sessions := session.NewSessionManager(
		time.Minute,  // GC every minute
		3*time.Hour,  // Expire idle after 3 hours
		12*time.Hour, // Expire absolute after 12 hours
		"session",    // Cookie name
		l,
	)

	// Read address from environment variable
	address := os.Getenv("PTT_ADDR")
	if address == "" {
		address = ":8080"
	}

	server.Start(&server.Config{
		Logger:   l,
		Static:   f,
		Address:  address,
		DB:       db,
		Sessions: sessions,
		DevMode:  devMode,
	})
}
