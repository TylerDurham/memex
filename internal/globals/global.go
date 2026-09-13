// Package globals
package globals

import "log/slog"

const AppName = "memex"
const DBName = AppName + "_index.db"

type App struct {
	Logger *slog.Logger
	Config ConfigInfo
}

func InitApp() (app App, err error) {
	app.Logger = newLogger()

	cfg, err := NewConfig()
	app.Config = cfg
	return app, err
}


