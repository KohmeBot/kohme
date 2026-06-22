package main

import (
	"github.com/kohmebot/kohme"
	"github.com/kohmebot/kohme/internal/app"
	"github.com/kohmebot/kohme/pkg/conf"
	"github.com/kohmebot/plugin/v2"
)

func main() {
	pgs := append([]plugin.Plugin{new(app.Core)}, kohme.GetPlugins()...)

	conf.ExportConfigSchemas(pgs)
}
