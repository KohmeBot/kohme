package testplugin

import (
	"github.com/kohmebot/plugin/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type MemPlugin struct {
	heap [][]byte
}

func (m *MemPlugin) OnInit(engine plugin.Engine, env plugin.Env) error {

	engine.OnCommand("alloc", env.SuperUser().Rule()).Handle(func(ctx *zero.Ctx) {

		b := make([]byte, 1024*1024)

		m.heap = append(m.heap, b)
		ctx.Send("已申请1MB堆内存")

	}).SetBlock(true)

	return nil
}

func (m *MemPlugin) OnBoot() {
	m.heap = make([][]byte, 0)
}

func (m *MemPlugin) OnHelp(ctx *zero.Ctx) {

}

func (m *MemPlugin) Name() string {
	return "mem"
}

func (m *MemPlugin) Version() string {
	return "v0.0.1"
}
