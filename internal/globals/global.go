package globals

import "log/slog"

const AppName = "memex"
const DbName = AppName + "_index.db"

type App struct {
	Logger *slog.Logger
	Config ConfigInfo
}

func InitApp() (app App, err error) {
	app.Logger = newLogger()

	cfg, err := newConfig()
	app.Config = cfg
	return app, err
}


