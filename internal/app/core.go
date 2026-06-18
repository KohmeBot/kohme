package app

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/kohmebot/kohme/internal/util"
	"github.com/kohmebot/pkg/chain"
	"github.com/kohmebot/pkg/command"
	"github.com/kohmebot/plugin/v2"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"
	"golang.org/x/mod/semver"
	"gorm.io/gorm"
	"runtime"
	"strings"
	"time"
)

const coreVersion = "v1.1.31"

type CoreConf struct {
	HelpTop  string `yaml:"help_top" jsonschema:"description=Help信息顶部"`
	HelpTail string `yaml:"help_tail" jsonschema:"description=Help信息尾部"`
}

type Core struct {
	app  *App
	conf CoreConf
	db   *gorm.DB
	env  plugin.Env
}

func newCore(a *App) *Core {
	return &Core{
		app: a,
	}
}

func (c *Core) ConfigModel() any {
	return new(CoreConf)
}

func (c *Core) OnInit(engine plugin.Engine, env plugin.Env) error {
	c.env = env
	err := env.GetConf(&c.conf)
	if err != nil {
		return err
	}
	c.db, err = env.GetDB()
	if err != nil {
		return err
	}
	err = c.db.AutoMigrate(&PluginRecord{})

	register := []func(plugin.Engine, plugin.Env) error{
		c.onHelp,
		c.onPing,
		c.onPlugin,
		c.onToggle,
		c.onMetric,
		c.onRestart,
	}

	for _, r := range register {
		err = r(engine, env)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Core) OnBoot() {
	var err error
	defer func() {
		if err != nil {
			logrus.Errorf("查询插件校验错误: %s", err.Error())
		}
	}()
	initPluginSet := mapset.NewSet[string]()
	initPluginSet.Append(c.app.pluginNameSeq...)
	var records []PluginRecord
	if err = c.db.Find(&records).Error; err != nil {
		return
	}
	historyPluginMp := make(map[string]PluginRecord, len(records))
	historyPluginSet := mapset.NewSet[string]()
	for _, record := range records {
		historyPluginMp[record.Name] = record
		historyPluginSet.Add(record.Name)
	}
	// 查看是否有新加载的插件
	var newPlugins []plugin.Plugin
	initPluginSet.Difference(historyPluginSet).Each(func(s string) bool {
		newPlugins = append(newPlugins, c.app.pluginMp[s])
		// 返回false才是继续迭代
		return false
	})

	// 查看是否有卸载的插件
	var deletePlugins []string // 卸载的插件只能用string表示
	historyPluginSet.Difference(initPluginSet).Each(func(s string) bool {
		deletePlugins = append(deletePlugins, s)
		return false
	})

	// 查看是否有版本变动的插件
	var updatePlugins []plugin.Plugin
	initPluginSet.Intersect(historyPluginSet).Each(func(s string) bool {
		r := historyPluginMp[s]

		if r.Version != c.app.pluginMp[s].Version() {
			updatePlugins = append(updatePlugins, c.app.pluginMp[s])
		}
		return false
	})
	err = c.db.Transaction(func(tx *gorm.DB) error {
		// 删除记录中已卸载的插件
		if len(deletePlugins) > 0 {
			if err = c.db.Where("name IN ?", deletePlugins).Delete(&PluginRecord{}).Error; err != nil {
				return err
			}
		}
		if len(newPlugins) > 0 {
			// 插入新的插件
			if err = c.db.Create(PluginsToRecord(newPlugins)).Error; err != nil {
				return err
			}
		}
		if len(updatePlugins) > 0 {
			// 更新插件版本
			for _, record := range PluginsToRecord(updatePlugins) {
				if err = c.db.Save(&record).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[%s]\nKohmeBot已启动\n", time.Now().Format("2006-01-02 15:04:05")))
	if len(c.app.pluginNameSeq) > 0 {
		builder.WriteString("已加载插件:\n")
		for idx, s := range c.app.pluginNameSeq {
			p := c.app.pluginMp[s]
			builder.WriteString(fmt.Sprintf("(%d) [%s] %s\n", idx+1, p.Name(), p.Version()))
		}
	}
	if len(newPlugins) > 0 {
		builder.WriteString("新插件:\n")
		for _, p := range newPlugins {
			builder.WriteString(fmt.Sprintf("[%s] %s\n", p.Name(), p.Version()))
		}
	}
	if len(deletePlugins) > 0 {
		builder.WriteString("卸载插件:\n")
		for _, s := range deletePlugins {
			r := historyPluginMp[s]
			builder.WriteString(fmt.Sprintf("[%s] %s\n", r.Name, r.Version))
		}
	}
	if len(updatePlugins) > 0 {
		builder.WriteString("版本变动:\n")
		for _, p := range updatePlugins {
			hp := historyPluginMp[p.Name()]
			var w string
			if semver.Compare(p.Version(), hp.Version) > 0 {
				w = "版本更新"
			} else {
				w = "版本回退"
			}
			builder.WriteString(fmt.Sprintf("[%s] %s %s -> %s\n", p.Name(), w, hp.Version, p.Version()))
		}
	}
	logrus.Info(builder.String())
	msg := message.Text(builder.String())

	c.env.UseBot(func(ctx *zero.Ctx) {
		for u := range c.env.SuperUser().RangeUser() {
			// 在OnBoot期间，ZeroBot的事件循环还未开始
			// 这里使用异步防止这里阻塞
			// 默认的一分钟时间足够撑到事件循环开始了
			go ctx.SendPrivateMessage(u, msg)
		}
	})

}

func (c *Core) OnHelp(ctx *zero.Ctx) {

	help := command.HelpTemplate{
		PluginName: "core",
		PluginDesc: "kohme 核心插件",
		Commands: []command.Command{
			{
				CMD: "help",
				Args: []command.Arg{
					{
						Name: "插件名称",
					},
				},
				Desc: "查看对应插件帮助",
			},
			{
				CMD:  "ping",
				Desc: "ping一下",
			},
			{
				CMD:  "plugin",
				Desc: "查看所有插件",
			},
			{
				CMD: "toggle",
				Args: []command.Arg{
					{
						Name: "插件名称",
					},
				},
				Desc: "开启/关闭插件",
			},
			{
				CMD: "metric",
				Args: []command.Arg{
					{
						Name:     "插件名称",
						Optional: true,
					},
				},
				Desc: "查看性能指标",
			},
			{
				CMD:  "restart",
				Desc: "重启Kohme",
			},
		},
	}

	ctx.Send(help.String())
}

func (c *Core) Name() string {
	return "core"
}

func (c *Core) Version() string {
	return coreVersion
}

func (c *Core) getEnvRule(env plugin.Env) zero.Rule {
	g := env.Groups()
	u := env.SuperUser()
	return func(ctx *zero.Ctx) bool {
		return g.Rule()(ctx) || u.Rule()(ctx)
	}
}

func (c *Core) onHelp(engine plugin.Engine, env plugin.Env) error {

	prefix := c.app.opt.AppConf.Zero.CommandPrefix
	engine.OnCommandGroup([]string{"help", "?", "？", "帮助"}, c.getEnvRule(env)).Handle(func(ctx *zero.Ctx) {
		var cmd extension.CommandModel
		err := ctx.Parse(&cmd)
		if err != nil {
			c.env.Error(ctx, err)
			return
		}
		cmd.Args = strings.TrimSpace(cmd.Args)
		if len(cmd.Args) > 0 {
			name := strings.Fields(cmd.Args)[0]
			p, ok := c.app.pluginMp[name]
			if !ok {
				c.env.Error(ctx, fmt.Errorf("插件 %s 不存在", name))
				return
			}
			pEnv := c.app.envMp[name]
			if pEnv.IsDisable() {
				return
			}
			if !c.getEnvRule(pEnv)(ctx) {
				return
			}
			p.OnHelp(ctx)
			return
		}

		var msgChain chain.MessageChain
		msgChain.Split(
			message.Text(c.conf.HelpTop),
			message.Text(fmt.Sprintf(`命令前缀 "%s"`, prefix)),
			message.Text(fmt.Sprintf("使用/help [插件名称] 查看插件详细帮助")),
		)
		msgChain.Line()
		for _, name := range c.app.pluginNameSeq {
			pEnv := c.app.envMp[name]
			// 跳过关闭的插件
			if pEnv.IsDisable() {
				continue
			}
			// 跳过未启用群的插件
			if !c.getEnvRule(pEnv)(ctx) {
				continue
			}

			p := c.app.pluginMp[name]
			msgChain.Line(message.Text(fmt.Sprintf("🌟%s", p.Name())))
		}
		msgChain.Split(message.Text("-----"), message.Text(c.conf.HelpTail))
		ctx.Send(msgChain)
	}).SetBlock(true)
	return nil
}

func (c *Core) onPing(engine plugin.Engine, env plugin.Env) error {
	supers := env.SuperUser()
	engine.OnCommand("ping", supers.Rule()).Handle(func(ctx *zero.Ctx) {
		ctx.Send(message.Text("pong!我还活着"))
	}).SetBlock(true)
	return nil
}

func (c *Core) onPlugin(engine plugin.Engine, env plugin.Env) error {
	supers := env.SuperUser()
	engine.OnCommand("plugin", supers.Rule()).Handle(func(ctx *zero.Ctx) {
		var msgChain chain.MessageChain
		msgChain.Line(message.Text("当前插件列表:"))
		for _, name := range c.app.pluginNameSeq {
			p := c.app.pluginMp[name]
			e := c.app.envMp[name]
			var toggle string
			disable := e.IsDisable()
			if disable {
				toggle = "关闭"
			} else {
				toggle = "开启"
			}
			msgChain.Join(message.Text(fmt.Sprintf("%s %s (%s)", p.Name(), p.Version(), toggle)))
			msgChain.Line()
		}
		ctx.Send(msgChain)
	}).SetBlock(true)
	return nil
}

func (c *Core) onToggle(engine plugin.Engine, env plugin.Env) error {
	supers := env.SuperUser()
	engine.OnCommand("toggle", supers.Rule()).Handle(func(ctx *zero.Ctx) {
		var cmd extension.CommandModel
		var err error
		defer func() {
			if err != nil {
				env.Error(ctx, err)
				return
			}
		}()
		err = ctx.Parse(&cmd)
		if err != nil {
			return
		}
		pluginName := cmd.Args
		pluginName = strings.TrimSpace(pluginName)
		if len(pluginName) <= 0 {
			err = fmt.Errorf("插件名称为空")
			return
		}

		if pluginName == "core" {
			err = fmt.Errorf("无法关闭core")
			return
		}

		e, ok := c.app.envMp[pluginName]
		if !ok {
			err = fmt.Errorf("插件%s不存在", pluginName)
			return
		}
		var msgChain chain.MessageChain

		if e.Disable.CompareAndSwap(true, false) {
			e.Metric.Start()
			msgChain.SplitEmpty(message.Text(pluginName), message.Text("已开启"))
		} else {
			e.Metric.Cleanup()
			e.Disable.CompareAndSwap(false, true)
			msgChain.SplitEmpty(message.Text(pluginName), message.Text("已关闭"))
		}
		ctx.Send(msgChain)

	}).SetBlock(true)
	return nil
}

func (c *Core) onMetric(engine plugin.Engine, env plugin.Env) error {
	supers := env.SuperUser()
	engine.OnCommand("metric", supers.Rule()).Handle(func(ctx *zero.Ctx) {
		var cmd extension.CommandModel
		err := ctx.Parse(&cmd)
		if err != nil {
			c.env.Error(ctx, err)
			return
		}
		cmd.Args = strings.TrimSpace(cmd.Args)
		if len(cmd.Args) <= 0 {
			cmd.Args = c.Name()
		}
		name := strings.Fields(cmd.Args)[0]
		_, ok := c.app.pluginMp[name]
		if !ok {
			c.env.Error(ctx, fmt.Errorf("插件 %s 不存在", name))
			return
		}
		pEnv := c.app.envMp[name]
		if pEnv.IsDisable() {
			return
		}
		if !c.getEnvRule(pEnv)(ctx) {
			return
		}

		var msgChain chain.MessageChain
		runtime.GC()
		if name == c.Name() {
			// core
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			msgChain.Line(message.Text(fmt.Sprintf("kohme占用内存: %.4fMB", float64(m.HeapAlloc)/1000/1000)))
		} else {
			mem, err := pEnv.HeapMemory()
			if err != nil {
				c.env.Error(ctx, err)
				return
			}
			msgChain.Line(message.Text(fmt.Sprintf("%s占用内存: %.4fMB", name, float64(mem)/1000/1000)))
		}

		msgChain.Line(message.Text(pEnv.MetricReport()))

		ctx.Send(msgChain)

	}).SetBlock(true)
	return nil
}

func (c *Core) onRestart(engine plugin.Engine, env plugin.Env) error {
	supers := env.SuperUser()
	engine.OnCommand("restart", supers.Rule()).Handle(func(ctx *zero.Ctx) {

		go func() {
			c.app.gate.Close()
			ch := make(chan struct{})
			go func() {
				c.app.gate.WaitFinish()
				close(ch)
			}()
			ctx.Send("正在等待所有插件处理完成...")
			var text string
			select {
			case <-ch:
				text = "正在重启kohme..."
			case <-time.After(10 * time.Second):
				text = "等待超时，将强制重启kohme"
			}
			time.Sleep(time.Second)
			ctx.Send(text)

			exit, err := util.Restart()
			if err != nil {
				env.Error(ctx, fmt.Errorf("重启失败,请尝试手动重启: %w", err))
				c.app.gate.Open()
				return
			}
			exit()
		}()

	}).SetBlock(true)
	return nil
}
