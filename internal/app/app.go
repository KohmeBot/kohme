package app

import (
	"fmt"
	fplugin "github.com/kohmebot/kohme/pkg/plugin"
	"github.com/kohmebot/pkg/version"
	"github.com/kohmebot/plugin"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type App struct {
	opt option

	Engine  *zero.Engine
	manager *fplugin.Manager

	pluginMp map[string]plugin.Plugin
	// 插件名称序列，这个表明了加载顺序
	pluginNameSeq []string
	envMp         map[string]*Env
}

func New(opts ...Option) *App {
	defaultOpt := defaultOption()
	for _, opt := range opts {
		opt(&defaultOpt)
	}

	a := &App{
		opt:      defaultOpt,
		Engine:   zero.New(),
		manager:  fplugin.NewPluginManager(defaultOpt.PluginConf.Path),
		pluginMp: make(map[string]plugin.Plugin),
		envMp:    make(map[string]*Env),
	}

	return a
}

func (a *App) Start() error {
	ps, err := a.manager.LoadPlugins()
	if err != nil {
		return err
	}
	a.RegisterPlugins(newCore(a))
	a.RegisterPlugins(append(a.opt.DefaultPlugins, ps...)...)

	for _, name := range a.pluginNameSeq {
		p := a.pluginMp[name]
		err = p.Init(a.Engine, a.envMp[p.Name()])
		if err != nil {
			return fmt.Errorf("%s 初始化失败: %w", p.Name(), err)
		}
	}
	a.PrintPlugins()
	zero.RunAndBlock(&a.opt.AppConf.Zero, func() {
		for _, name := range a.pluginNameSeq {
			a.pluginMp[name].OnBoot()
		}
	})
	return nil

}

// RegisterPlugins 注册插件
func (a *App) RegisterPlugins(pluginsToAdd ...plugin.Plugin) {
	// 获取现有插件配置
	pluginsConfig := a.opt.PluginConf.Plugins

	// 过滤掉不需要加载的
	filteredPlugins := pluginsConfig.filterInvalidPlugins(pluginsToAdd)
	// 为加载顺序排序
	pluginsConfig.sortPluginsBySequence(filteredPlugins)

	// 插入并配置插件
	for _, p := range filteredPlugins {
		_, ok := pluginsConfig[p.Name()]
		if !ok {
			logrus.Warnf("插件 %s 的配置不存在,将使用默认配置", p.Name())
		}
		a.addPlugin(p)
	}
}

// 添加并配置单个插件
func (a *App) addPlugin(p plugin.Plugin) {
	// 检查插件是否已存在
	if _, exists := a.pluginMp[p.Name()]; exists {
		panic(fmt.Sprintf("存在重复名称的插件: %s", p.Name()))
	}

	// 插件注册
	a.pluginMp[p.Name()] = p
	a.pluginNameSeq = append(a.pluginNameSeq, p.Name())

	// 配置插件环境
	a.envMp[p.Name()] = a.configurePluginEnv(p)
}

// 配置插件环境
func (a *App) configurePluginEnv(p plugin.Plugin) *Env {
	customConf := a.opt.PluginConf.Plugins[p.Name()]
	if customConf.Conf == nil {
		customConf.Conf = make(map[string]any)
	}
	if customConf.Other == nil {
		customConf.Other = make(map[string]any)
	}
	// 如果不存在插件自定配置启用的群，则使用全局配置
	if len(customConf.Groups) <= 0 {
		customConf.Groups = a.opt.PluginConf.Groups
	}
	// 同理
	if len(customConf.SuperUsers) <= 0 {
		customConf.SuperUsers = a.opt.AppConf.Zero.SuperUsers
	}
	return NewEnv(p, customConf, a.pluginMp)
}

func (a *App) PrintPlugins() {
	for _, name := range a.pluginNameSeq {
		p := a.pluginMp[name]
		logrus.Infof("插件 %s | 版本 %s | 描述 %s", p.Name(), version.Version(p.Version()), p.Description())
	}
}
