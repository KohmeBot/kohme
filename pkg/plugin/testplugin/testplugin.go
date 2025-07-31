package testplugin

import (
	"fmt"
	"github.com/kohmebot/pkg/command"
	"github.com/kohmebot/pkg/version"
	"github.com/kohmebot/plugin"
	zero "github.com/wdvxdr1123/ZeroBot"
	"time"
)

// TestPlugin 用于测试的插件,返回any error
type TestPlugin struct {
	ErrorDuration time.Duration
	env           plugin.Env
}

func (t *TestPlugin) Init(engine *zero.Engine, env plugin.Env) error {
	t.env = env
	engine.OnMessage(env.SuperUser().Rule()).Handle(func(ctx *zero.Ctx) {
		env.Error(ctx, fmt.Errorf("test error"))
	})

	return nil
}

func (t *TestPlugin) Name() string {
	return "test"
}

func (t *TestPlugin) Description() string {
	return "test plugin"
}

func (t *TestPlugin) Commands() command.Commands {
	return command.NewCommands()
}

func (t *TestPlugin) Version() version.Version {
	return 0
}

func (t *TestPlugin) OnBoot() {
	env := t.env
	if t.ErrorDuration > 0 {
		ticker := time.NewTicker(t.ErrorDuration)
		go func() {
			defer ticker.Stop()
			for range ticker.C {
				for ctx := range env.RangeBot {
					ctx.Event = &zero.Event{}
					for gid := range env.Groups().RangeGroup {
						ctx.Event.GroupID = gid
						env.Error(ctx, fmt.Errorf("test error tock %d", gid))
					}
				}
			}
		}()
	}
}
