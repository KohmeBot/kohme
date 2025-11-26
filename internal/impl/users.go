package impl

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"iter"
)

type Users []int64

func (u Users) IsContains(userId int64) bool {
	for _, user := range u {
		if user == userId {
			return true
		}
	}
	return false
}

func (u Users) Rule() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		userId := ctx.Event.Sender.ID
		if userId <= 0 {
			return false
		}

		return u.IsContains(userId)
	}
}

func (u Users) RangeUser() iter.Seq[int64] {
	return func(yield func(int64) bool) {
		for _, user := range u {
			if !yield(user) {
				return
			}
		}
	}
}

type UserWithEnv struct {
	Users
	env *Env
}

func NewUserWithEnv(users Users, env *Env) *UserWithEnv {
	return &UserWithEnv{
		users, env,
	}
}

func (g *UserWithEnv) Rule() zero.Rule {
	rule := g.Users.Rule()
	return func(ctx *zero.Ctx) bool {
		if g.env.IsDisable() {
			return false
		}
		return rule(ctx)
	}
}
