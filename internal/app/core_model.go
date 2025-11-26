package app

import "github.com/kohmebot/plugin/v2"

// PluginRecord 插件记录
type PluginRecord struct {
	Name    string `gorm:"primaryKey"`
	Version string
}

func PluginsToRecord(ps []plugin.Plugin) []PluginRecord {
	rs := make([]PluginRecord, len(ps))
	for i, p := range ps {
		rs[i].Name = p.Name()
		rs[i].Version = p.Version()
	}
	return rs
}
