package main

import (
	"fmt"
	"github.com/kohmebot/kohme"
	"github.com/kohmebot/kohme/internal/app"
	"github.com/kohmebot/kohme/pkg/conf"
	"github.com/kohmebot/plugin/v2"
	"os"
	"reflect"
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

func register(v any) {
	var p plugin.Plugin
	switch t := v.(type) {
	case func() plugin.Plugin:
		p = t()
	case plugin.Plugin:
		p = t
	default:
		rv := reflect.ValueOf(v)

		if rv.Kind() != reflect.Func {
			break
		}
		res := rv.Call([]reflect.Value{})
		for _, rv = range res {
			if !rv.CanInterface() || rv.IsNil() {
				continue
			}
			var ok bool
			p, ok = rv.Interface().(plugin.Plugin)
			if ok {
				break
			}
		}
	}

	if p == nil {
		panic(fmt.Errorf("can not register %T", v))
	}

	kohme.Register(p)
}
