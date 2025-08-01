//go:generate go run ./cmd/plugin_gen
//go:generate go fmt ./cmd/bot/plugin.gen.go
//go:generate go mod tidy
package kohme

import "github.com/kohmebot/plugin"

var plugins []plugin.Plugin

func Register(f func() plugin.Plugin) {
	plugins = append(plugins, f())
}

func GetPlugins() []plugin.Plugin {
	return plugins
}
