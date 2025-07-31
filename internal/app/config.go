package app

import (
	"encoding/json"
	"fmt"
	"github.com/kohmebot/plugin"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"gopkg.in/yaml.v3"
	"os"
	"slices"
)

type WsConf struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type AConf struct {
	Zero      zero.Config `json:"zero"`
	Ws        WsConf      `json:"ws"`
	ReverseWs WsConf      `json:"rws"`
}

func (c *AConf) InitDriver() {
	var ds []zero.Driver
	if c.Ws.Url != "" {
		// 正向Ws
		ds = append(ds, driver.NewWebSocketClient(c.Ws.Url, c.Ws.Token))
	}
	if c.ReverseWs.Url != "" {
		// 反向Ws
		ds = append(ds, driver.NewWebSocketServer(16, c.ReverseWs.Url, c.ReverseWs.Token))
	}
	c.Zero.Driver = ds
}

func (c *AConf) ParseJsonFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无法打开 JSON 文件: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(c); err != nil {
		return fmt.Errorf("解析 JSON 文件错误: %w", err)
	}

	c.InitDriver()

	return nil
}

// PluginConf 对应plugins.yaml
type PluginConf struct {
	Path    string        `yaml:"path"`
	Plugins PluginConfMap `yaml:"plugins"`
	Groups  []int64       `yaml:"groups"`
}

func (c *PluginConf) ParseYamlFile(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("无法读取 YAML 文件: %w", err)
	}

	if err := yaml.Unmarshal(file, c); err != nil {
		return fmt.Errorf("解析 YAML 文件错误: %w", err)
	}

	return nil
}

// CustomPluginConf 自定义插件配置
type CustomPluginConf struct {
	// 插件仓库地址
	Repo string `yaml:"repo"`
	// 插件指定版本
	Version string `yaml:"version"`
	// 决定加载顺序
	Seq int64 `yaml:"seq"`
	// 是否排除(不加载)
	Exclude bool `yaml:"exclude"`
	// 是否禁用功能(但依旧加载)
	Disable bool `yaml:"disable"`
	// 开启的群组
	Groups []int64 `yaml:"groups"`
	// 超级管理员列表
	SuperUsers []int64 `yaml:"super_users"`
	// 插件自定义conf
	Conf map[string]any `yaml:"conf"`
	// 其他不定字段,作为环境变量传入
	Other map[string]any `yaml:",inline"`
}

// PluginConfMap 插件配置映射，key为插件名称
type PluginConfMap map[string]CustomPluginConf

// 过滤不需要加载的插件
func (mp PluginConfMap) filterInvalidPlugins(plugins []plugin.Plugin) []plugin.Plugin {
	return slices.DeleteFunc(plugins, func(p plugin.Plugin) bool {
		return mp[p.Name()].Exclude
	})
}

// 根据顺序排序插件
func (mp PluginConfMap) sortPluginsBySequence(plugins []plugin.Plugin) {
	slices.SortFunc(plugins, func(a, b plugin.Plugin) int {
		aSeq := mp[a.Name()].Seq
		bSeq := mp[b.Name()].Seq
		return int(aSeq - bSeq)
	})
}
