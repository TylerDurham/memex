package globals

type AppInfo struct {
	name    string
	dbName  string
	version string
}

func (a *AppInfo) Name() string {
	return a.name
}

func (a *AppInfo) DBName() string {
	return a.dbName
}

func (a *AppInfo) Version() string {
	return a.version
}

var app *AppInfo

func App() *AppInfo {
	if app == nil {
		app = &AppInfo{
			name:    "memex",
			dbName:  "memex_store.db",
			version: "TBD",
		}
	}
	return app
}

// type AppServicesInfo struct {
// 	info   AppInfo
// 	logger *slog.Logger
// }
//
// func (a *AppServicesInfo) Log() *slog.Logger {
// 	return a.logger
// }
//
//
// func AppServices() (*AppServicesInfo, error) {
//
// 	var app = &AppServicesInfo{
// 		info:   *newAppInfo(),
// 	}
//
// 	return app, nil
// }
