package impl

import (
	"fmt"
	"github.com/kohmebot/kohme/internal/db"
	"github.com/kohmebot/kohme/pkg/conf"
	"github.com/kohmebot/pkg/chain"
	"github.com/kohmebot/plugin/v2"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"sync/atomic"
)

type Env struct {
	customConf   conf.CustomPluginConf
	p            plugin.Plugin
	otherPlugins map[string]plugin.Plugin
	Disable      atomic.Bool
	superUser    Users
	group        *GroupsWithEnv
	envs         map[string]any
}

func NewEnv(p plugin.Plugin, customConf conf.CustomPluginConf, otherPlugins map[string]plugin.Plugin, envs map[string]any) *Env {
	e := &Env{
		p:            p,
		customConf:   customConf,
		otherPlugins: otherPlugins,
		envs:         envs,
	}
	e.Disable.Store(customConf.Disable)
	e.superUser = customConf.SuperUsers
	e.group = NewGroupsWithEnv(customConf.Groups, e)
	return e
}

func (e *Env) Groups() plugin.Groups {
	return e.group
}

func (e *Env) SuperUser() plugin.Users {
	return e.superUser
}

func (e *Env) Error(ctx *zero.Ctx, err error) {
	if err == nil {
		return
	}
	logrus.Errorf(fmt.Sprintf("[%s] %v", e.p.Name(), err))
	var msgChain chain.MessageChain

	sendToSuperUsers := func() {
		for user := range e.superUser.RangeUser() {
			ctx.SendPrivateMessage(user, msgChain)
		}
	}

	send := sendToSuperUsers // 默认情况下发送给超级用户

	if ctx.Event != nil {
		if zero.OnlyGroup(ctx) {
			// 在群聊中需要reply
			msgId := ctx.Event.MessageID
			msgChain.Join(message.Reply(msgId))
		}

		if ctx.Event.GroupID > 0 {
			gid := ctx.Event.GroupID
			send = func() {
				ctx.SendGroupMessage(gid, msgChain)
			}
		} else if ctx.Event.UserID > 0 {
			uid := ctx.Event.UserID
			send = func() {
				ctx.SendPrivateMessage(uid, msgChain)
			}
		}
	}

	msgChain.Split(
		message.Text(fmt.Sprintf("Oops！%s发生错误了！", e.p.Name())),
		message.Text(err.Error()),
	)

	send()
}

func (e *Env) Set(key string, value any) {
	e.envs[key] = value
}

func (e *Env) Get(key string) any {
	return e.envs[key]
}

func (e *Env) FilePath() (string, error) {
	path := filepath.Join("data", e.p.Name())
	err := os.MkdirAll(path, os.ModePerm)
	return path, err
}

func (e *Env) UseBot(h zero.Handler) {
	if e.IsDisable() {
		return
	}
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		h(ctx)
		return false
	})
}

func (e *Env) GetConf(conf any) error {
	data, err := yaml.Marshal(e.customConf.Conf)
	if err != nil {
		return fmt.Errorf("解析配置错误: %v", err)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return fmt.Errorf("解析配置错误: %v", err)
	}
	return nil
}

func (e *Env) GetDB() (*gorm.DB, error) {
	p, err := e.FilePath()
	if err != nil {
		return nil, err
	}
	return db.Get(filepath.Join(p, fmt.Sprintf("%s.db", e.p.Name())))
}

func (e *Env) GetPlugin(name string) (p plugin.Plugin, ok bool) {
	p, ok = e.otherPlugins[name]
	return
}

// IsDisable 判断是否禁用
func (e *Env) IsDisable() bool {
	return e.Disable.Load()
}

func (e *Env) Toggle(b bool) {
	e.Disable.Store(!b)
}
