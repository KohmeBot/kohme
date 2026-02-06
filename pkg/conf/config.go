package conf

import (
	"encoding/json"
	"fmt"
	"github.com/kohmebot/plugin/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type WsConf struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type ZeroConf struct {
	Zero      zero.Config `json:"zero"`
	Ws        WsConf      `json:"ws"`
	ReverseWs WsConf      `json:"rws"`
}

func (c *ZeroConf) initDriver() {
	var ds []zero.Driver
	if c.Ws.Url != "" {
		// 正向Ws
		clear(ds)
		ds = append(ds, driver.NewWebSocketClient(c.Ws.Url, c.Ws.Token))
	}
	if c.ReverseWs.Url != "" {
		// 反向Ws
		clear(ds)
		ds = append(ds, driver.NewWebSocketServer(16, c.ReverseWs.Url, c.ReverseWs.Token))
	}
	c.Zero.Driver = ds
}

func (c *ZeroConf) ParseJsonFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无法打开 JSON 文件: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(c); err != nil {
		return fmt.Errorf("解析 JSON 文件错误: %w", err)
	}

	c.initDriver()

	return nil
}

// PluginConf 对应plugins.yaml
type PluginConf struct {
	Path    string        `yaml:"path"`
	Plugins PluginConfMap `yaml:"plugins"`
	Groups  []int64       `yaml:"groups"`
	// 环境变量
	Other map[string]any `yaml:"env"`
}

func (c *PluginConf) ParseYamlFile(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("无法读取 YAML 文件: %w", err)
	}

	if err := yaml.Unmarshal(file, c); err != nil {
		return fmt.Errorf("解析 YAML 文件错误: %w", err)
	}

	if len(c.Path) == 0 {
		c.Path = PluginPath
	}

	err = c.parseFromDir(c.Path)
	if err != nil {
		return err
	}

	for name, conf := range c.Plugins {
		if len(conf.Repo) <= 0 {
			conf.Repo = fmt.Sprintf("github.com/kohmebot/%s", name)
		}
	}
	return nil

}

func (c *PluginConf) parseFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("无法读取目录: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			fullPath := filepath.Join(dir, name)

			if fullPath == filepath.Clean(PluginConfigPath) {
				continue
			}
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return fmt.Errorf("无法读取文件 %s: %w", fullPath, err)
			}

			if err := yaml.Unmarshal(data, c.Plugins); err != nil {
				return fmt.Errorf("解析 YAML 文件 %s 错误: %w", fullPath, err)
			}
		}
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
}

// PluginConfMap 插件配置映射，key为插件名称
type PluginConfMap map[string]CustomPluginConf

// FilterInvalid 过滤不需要加载的插件
func (mp PluginConfMap) FilterInvalid(plugins []plugin.Plugin) []plugin.Plugin {
	return slices.DeleteFunc(plugins, func(p plugin.Plugin) bool {
		return mp[p.Name()].Exclude
	})
}

// SortBySequence 根据顺序排序插件
func (mp PluginConfMap) SortBySequence(plugins []plugin.Plugin) {
	slices.SortFunc(plugins, func(a, b plugin.Plugin) int {
		aSeq := mp[a.Name()].Seq
		bSeq := mp[b.Name()].Seq
		return int(aSeq - bSeq)
	})
}
