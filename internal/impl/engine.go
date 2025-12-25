package impl

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

type EnvEngine struct {
	env *Env
	e   *zero.Engine
}

func NewEngine(env *Env, e *zero.Engine) *EnvEngine {
	return &EnvEngine{
		env: env,
		e:   e,
	}
}

func (e *EnvEngine) withEnableRule(rules []zero.Rule) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if e.env.IsDisable() {
			return false
		}
		for _, rule := range rules {
			if !rule(ctx) {
				return false
			}
		}
		return true
	}
}

func (e *EnvEngine) wrapEnableRule(rules []zero.Rule) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if e.env.IsDisable() {
			return true
		}
		for _, rule := range rules {
			if !rule(ctx) {
				return false
			}
		}
		return true
	}
}

func (e *EnvEngine) wrapEnableHandler(handler []zero.Handler) zero.Handler {
	return func(ctx *zero.Ctx) {
		if e.env.IsDisable() {
			return
		}
		for _, h := range handler {
			h(ctx)
		}
	}
}

func (e *EnvEngine) UsePreHandler(rules ...zero.Rule) {
	e.e.UsePreHandler(e.wrapEnableRule(rules))
}

func (e *EnvEngine) UseMidHandler(rules ...zero.Rule) {
	e.e.UseMidHandler(e.wrapEnableRule(rules))
}

func (e *EnvEngine) UsePostHandler(handler ...zero.Handler) {
	e.e.UsePostHandler(e.wrapEnableHandler(handler), func(ctx *zero.Ctx) {
		e.env.Metric.CommandEnd(string(ctx.Event.RawMessageID))
	})
}

func (e *EnvEngine) On(typ string, rules ...zero.Rule) *zero.Matcher {
	return e.e.On(typ, e.withEnableRule(rules))
}

func (e *EnvEngine) OnMessage(rules ...zero.Rule) *zero.Matcher {
	return e.e.OnMessage(e.withEnableRule(rules))
}

func (e *EnvEngine) OnNotice(rules ...zero.Rule) *zero.Matcher {
	return e.e.OnNotice(e.withEnableRule(rules))
}

func (e *EnvEngine) OnRequest(rules ...zero.Rule) *zero.Matcher {
	return e.e.OnRequest(e.withEnableRule(rules))
}

func (e *EnvEngine) OnMetaEvent(rules ...zero.Rule) *zero.Matcher {
	return e.e.OnMetaEvent(e.withEnableRule(rules))
}

func (e *EnvEngine) OnPrefix(prefix string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnPrefix(prefix, e.withEnableRule(rules))
}

func (e *EnvEngine) OnSuffix(suffix string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnSuffix(suffix, e.withEnableRule(rules))
}

func (e *EnvEngine) OnCommand(commands string, rules ...zero.Rule) *zero.Matcher {
	command := commands
	return e.e.OnCommand(commands, e.withEnableRule(rules), func(ctx *zero.Ctx) bool {
		e.env.Metric.CommandStart(string(ctx.Event.RawMessageID), command)
		return true
	})
}

func (e *EnvEngine) OnRegex(regexPattern string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnRegex(regexPattern, e.withEnableRule(rules))
}

func (e *EnvEngine) OnKeyword(keyword string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnKeyword(keyword, e.withEnableRule(rules))
}

func (e *EnvEngine) OnFullMatch(src string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnFullMatch(src, e.withEnableRule(rules))
}

func (e *EnvEngine) OnFullMatchGroup(src []string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnFullMatchGroup(src, e.withEnableRule(rules))
}

func (e *EnvEngine) OnKeywordGroup(keywords []string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnKeywordGroup(keywords, e.withEnableRule(rules))
}

func (e *EnvEngine) OnCommandGroup(commands []string, rules ...zero.Rule) *zero.Matcher {
	var command string
	for _, c := range commands {
		command = c
		e.env.Metric.CommandInit(command)
		break
	}
	return e.e.OnCommandGroup(commands, e.withEnableRule(rules), func(ctx *zero.Ctx) bool {
		e.env.Metric.CommandStart(string(ctx.Event.RawMessageID), command)
		return true
	})
}

func (e *EnvEngine) OnPrefixGroup(prefix []string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnPrefixGroup(prefix, e.withEnableRule(rules))
}

func (e *EnvEngine) OnSuffixGroup(suffix []string, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnSuffixGroup(suffix, e.withEnableRule(rules))
}

func (e *EnvEngine) OnShell(command string, model interface{}, rules ...zero.Rule) *zero.Matcher {
	return e.e.OnShell(command, model, e.withEnableRule(rules))
}
