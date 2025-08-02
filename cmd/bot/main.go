package main

import (
	"fmt"
	"github.com/kohmebot/kohme"
	"github.com/kohmebot/kohme/internal/app"
	"github.com/kohmebot/kohme/pkg/conf"
	"github.com/kohmebot/plugin"
	"reflect"
)

func main() {
	aConf := conf.ZeroConf{}

	err := aConf.ParseJsonFile(conf.BotConfigPath)
	if err != nil {
		panic(err)
	}

	pluginConf := conf.PluginConf{}

	err = pluginConf.ParseYamlFile(conf.PluginConfigPath)
	if err != nil {
		panic(err)
	}

	a := app.New(
		app.WithAppConf(aConf),
		app.WithPluginConf(pluginConf),
		app.WithPlugin(kohme.GetPlugins()...),
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
