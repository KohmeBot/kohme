package main

import (
	"github.com/kohmebot/kohme"
	"github.com/kohmebot/kohme/internal/app"
	"github.com/kohmebot/kohme/pkg/conf"
	"os"
)

func main() {
	aConf := conf.ZeroConf{}

	err := aConf.ParseJsonFile(conf.BotConfigPath)
	if os.IsNotExist(err) {
		_ = conf.CreateZeroConf()
		err = aConf.ParseJsonFile(conf.BotConfigPath)
	}
	if err != nil {
		panic(err)
	}

	pluginConf := conf.PluginConf{}

	err = pluginConf.ParseYamlFile(conf.PluginConfigPath)
	if os.IsNotExist(err) {
		_ = conf.CreatePluginConf()
		err = pluginConf.ParseYamlFile(conf.PluginConfigPath)
	}
	if err != nil {
		panic(err)
	}

	pgs := kohme.GetPlugins()

	a := app.New(
		app.WithAppConf(aConf),
		app.WithPluginConf(pluginConf),
		app.WithPlugin(pgs...),
	)

	panic(a.Start())
}
