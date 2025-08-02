package kohme

import (
	"github.com/kohmebot/kohme/internal/app"
	"github.com/kohmebot/kohme/pkg/conf"
)

func RunKohme(zConf conf.ZeroConf, pConf conf.PluginConf) error {
	a := app.New(
		app.WithAppConf(zConf),
		app.WithPluginConf(pConf),
		app.WithPlugin(GetPlugins()...),
	)
	return a.Start()
}
