package main

import (
	"fmt"
	"github.com/kohmebot/kohme/internal/app"
	"github.com/kohmebot/kohme/pkg/conf"
	"os"
	"os/exec"
	"path"
	"strings"
)

const template = `
package main

import (
	%s
	"github.com/kohmebot/plugin"
)

// plugins 加载插件
func plugins() []plugin.Plugin {
	return []plugin.Plugin{
		%s
}
}
`

func main() {
	pluginConf := app.PluginConf{}

	err := pluginConf.ParseYamlFile(conf.PluginConfigPath)
	if err != nil {
		panic(err)
	}

	var (
		importBuilder strings.Builder
		newBuilder    strings.Builder
		getUrl        []string
	)

	for name, c := range pluginConf.Plugins {
		if len(c.Repo) <= 0 {
			c.Repo = fmt.Sprintf("github.com/kohmebot/%s", name)
		}
		name = path.Base(c.Repo)
		if name == "core" {
			continue
		}

		importBuilder.WriteString(fmt.Sprintf(`"%s"`, path.Join(c.Repo, name)))
		importBuilder.WriteByte('\n')
		newBuilder.WriteString(fmt.Sprintf("%s.NewPlugin(),", name))
		strings.TrimPrefix(c.Version, "v")
		if len(c.Version) <= 0 {
			getUrl = append(getUrl, fmt.Sprintf("%s@latest", c.Repo))
		} else {
			getUrl = append(getUrl, fmt.Sprintf("%s@v%s", c.Repo, c.Version))
		}

	}

	p := fmt.Sprintf(template, importBuilder.String(), newBuilder.String())

	err = writePlugin([]byte(p))
	if err != nil {
		panic(err)
	}

	for _, u := range getUrl {
		err = getMod(u)
		if err != nil {
			panic(err)
		}
	}

	err = tidy()
	if err != nil {
		panic(err)
	}

}

func writePlugin(b []byte) error {
	file, err := os.Create("./cmd/bot/plugin.go")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(b)
	return err
}

func getMod(url string) error {
	cmd := exec.Command("go", "get", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func tidy() error {
	// run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
