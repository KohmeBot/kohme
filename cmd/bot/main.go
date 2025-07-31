package main

import (
	"github.com/kohmebot/kohme/internal/app"
	"github.com/kohmebot/kohme/pkg/conf"
	"github.com/kohmebot/plugin"
)

var defaultPlugins []plugin.Plugin

func main() {
	aConf := app.AConf{}

	err := aConf.ParseJsonFile(conf.BotConfigPath)
	if err != nil {
		panic(err)
	}

	pluginConf := app.PluginConf{}

	err = pluginConf.ParseYamlFile(conf.PluginConfigPath)
	if err != nil {
		panic(err)
	}

	a := app.New(
		app.WithAppConf(aConf),
		app.WithPluginConf(pluginConf),
		app.WithPlugin(defaultPlugins...),
	)
	panic(a.Start())
}

func init() {
	defaultPlugins = plugins()
}
