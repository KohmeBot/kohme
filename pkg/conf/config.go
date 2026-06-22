package conf

import (
	"encoding/json"
	"fmt"
	"github.com/kohmebot/kohme/internal/util"
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
	Zero      zero.Config `json:"zero" jsonschema:"description=ZeroBot配置"`
	Ws        WsConf      `json:"ws" jsonschema:"description=正向WS配置"`
	ReverseWs WsConf      `json:"rws" jsonschema:"description=反向WS配置"`
}

func CreateZeroConf() error {
	c := &ZeroConf{
		Ws: WsConf{
			Url:   "ws://127.0.0.1:3001",
			Token: "",
		},
		ReverseWs: WsConf{
			Url:   "ws://127.0.0.1:3002",
			Token: "",
		},
		Zero: zero.Config{
			NickName:      []string{"kohme"},
			SuperUsers:    []int64{},
			CommandPrefix: "/",
		},
	}
	data, marshalErr := json.MarshalIndent(c, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("生成默认 JSON 配置失败: %w", marshalErr)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(BotConfigPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	if err := os.WriteFile(BotConfigPath, data, 0644); err != nil {
		return fmt.Errorf("写入默认配置失败: %w", err)
	}

	return nil
}

func (c *ZeroConf) initDriver() {
	var ds []zero.Driver
	if c.Ws.Url != "" {
		// 正向Ws
		ds = nil
		ds = append(ds, driver.NewWebSocketClient(c.Ws.Url, c.Ws.Token))
	}
	if c.ReverseWs.Url != "" {
		// 反向Ws
		ds = nil
		ds = append(ds, driver.NewWebSocketServer(16, c.ReverseWs.Url, c.ReverseWs.Token))
	}
	c.Zero.Driver = ds
}

func (c *ZeroConf) ParseJsonFile(path string) error {
	if !util.PathExists(path) {
		return os.ErrNotExist
	}

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
	Groups []int64 `yaml:"groups" jsonschema:"description=开启的群"`
	// 环境变量
	Other   map[string]any `yaml:"env" jsonschema:"description=环境变量"`
	Plugins PluginConfMap  `yaml:"plugins" jsonschema:"-"`
}

func CreatePluginConf() error {
	c := &PluginConf{
		Groups: []int64{},
		Plugins: PluginConfMap{
			"core": {
				Conf: map[string]any{
					"help_top":  "下面是我的所有本领！",
					"help_tail": "更多本领绝赞学习中,加入github.com/KohmeBot来教会我吧！",
				},
			},
		},
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("序列化 YAML 失败: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(PluginConfigPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(PluginConfigPath, data, 0644); err != nil {
		return fmt.Errorf("保存 YAML 失败: %w", err)
	}

	return nil

}

func (c *PluginConf) ParseYamlFile(path string) error {
	if !util.PathExists(path) {
		return os.ErrNotExist
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("无法读取 YAML 文件: %w", err)
	}

	if err := yaml.Unmarshal(file, c); err != nil {
		return fmt.Errorf("解析 YAML 文件错误: %w", err)
	}

	err = c.parseFromDir(ConfigPath)
	if err != nil {
		return err
	}

	for name, conf := range c.Plugins {
		if len(conf.Repo) <= 0 {
			conf.Repo = fmt.Sprintf("github.com/kohmebot/%s", name)
		}
		c.Plugins[name] = conf
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
	Repo string `yaml:"repo" jsonschema:"description=插件仓库地址"`
	// 插件指定版本
	Version string `yaml:"version" jsonschema:"description=插件版本"`
	// 决定加载顺序
	Seq int64 `yaml:"seq" jsonschema:"description=插件加载顺序"`
	// 是否排除(不加载)
	Exclude bool `yaml:"exclude" jsonschema:"description=不加载该插件"`
	// 是否禁用功能(但依旧加载)
	Disable bool `yaml:"disable" jsonschema:"description=禁用功能"`
	// 开启的群组
	Groups []int64 `yaml:"groups" jsonschema:"description=该插件开启的群组，留空则使用全局配置"`
	// 超级管理员列表
	SuperUsers []int64 `yaml:"super_users" jsonschema:"description=管理员列表，留空则使用全局配置"`
	// 插件自定义conf
	Conf map[string]any `yaml:"conf" jsonschema:"-"`
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
