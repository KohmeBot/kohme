package plugin

import "github.com/kohmebot/plugin/v2"

type Filter func(plugin.Plugin) bool

func NameFilter(names ...string) Filter {
	mp := map[string]struct{}{}
	for _, name := range names {
		mp[name] = struct{}{}
	}
	return func(plugin plugin.Plugin) bool {
		_, ok := mp[plugin.Name()]
		return !ok
	}
}
