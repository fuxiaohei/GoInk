package goink

import "github.com/fuxiaohei/goink/app"

const (
	VERSION = "0.1.1"
)

// create new app in top level
func NewApp(configFile string) *app.InkApp {
	return app.NewApp(configFile, app.CONFIG_JSON)
}

