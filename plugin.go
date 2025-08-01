//go:generate go run ./cmd/plugin_gen
//go:generate go fmt ./cmd/bot/plugin.gen.go
//go:generate go mod tidy
package kohme

import "github.com/kohmebot/plugin"

var plugins []plugin.Plugin

func Register(p plugin.Plugin) {
	plugins = append(plugins, p)
}

func GetPlugins() []plugin.Plugin {
	return plugins
}
