package impl

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"iter"
)

type Groups []int64

func (g Groups) RangeGroup() iter.Seq[int64] {
	return func(yield func(int64) bool) {
		for _, group := range g {
			if !yield(group) {
				return
			}
		}
	}
}

func (g Groups) IsContains(groupId int64) bool {
	for _, group := range g {
		if group == groupId {
			return true
		}
	}
	return false
}

func (g Groups) Rule() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		groupId := ctx.Event.GroupID
		if groupId <= 0 {
			return false
		}

		return g.IsContains(groupId)
	}
}

type GroupsWithEnv struct {
	Groups
	env *Env
}

func NewGroupsWithEnv(groups Groups, env *Env) *GroupsWithEnv {
	return &GroupsWithEnv{
		groups, env,
	}
}

func (g *GroupsWithEnv) Rule() zero.Rule {
	rule := g.Groups.Rule()
	return func(ctx *zero.Ctx) bool {
		if g.env.IsDisable() {
			return false
		}
		return rule(ctx)
	}
}
