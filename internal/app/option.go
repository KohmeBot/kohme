package app

import (
	"github.com/kohmebot/kohme/pkg/conf"
	"github.com/kohmebot/plugin"
)

type Option func(opt *option)

type option struct {
	PluginConf     conf.PluginConf
	AppConf        conf.ZeroConf
	DefaultPlugins []plugin.Plugin
}

func WithPlugin(p ...plugin.Plugin) Option {
	return func(opt *option) {
		opt.DefaultPlugins = append(opt.DefaultPlugins, p...)
	}
}

func WithPluginConf(conf conf.PluginConf) Option {
	return func(opt *option) {
		opt.PluginConf = conf
	}
}

func WithAppConf(conf conf.ZeroConf) Option {
	return func(opt *option) {
		opt.AppConf = conf
	}
}

func defaultOption() option {
	return option{
		PluginConf:     conf.PluginConf{},
		AppConf:        conf.ZeroConf{},
		DefaultPlugins: []plugin.Plugin{},
	}
}
