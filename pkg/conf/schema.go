package conf

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	"github.com/kohmebot/plugin/v2"
	"log"
	"os"
	"path/filepath"
)

// ExportConfigSchemas writes {pluginName: schema} to <confDir>/.schemas.json
// for every registered plugin that implements ConfigSchemaProvider.
func ExportConfigSchemas(pgs []plugin.Plugin) {
	out := map[string]json.RawMessage{}
	for _, p := range pgs {
		cp, ok := p.(plugin.ConfigProvider)
		if !ok {
			continue
		}
		model := cp.ConfigModel()
		if model == nil {
			continue
		}
		r := &jsonschema.Reflector{
			FieldNameTag: "yaml",
		}
		s := r.Reflect(model)
		schema, _ := s.MarshalJSON()

		// Validate it's real JSON so a typo in a plugin can't corrupt the file.
		if !json.Valid(schema) {
			log.Printf("ConfigSchema 不是合法 JSON，已跳过该插件")
			continue
		}
		out[p.Name()] = schema
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return
	}
	path := filepath.Join(ConfigPath, ".schemas.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		log.Printf("写入 %s 失败: %v", path, err)
	}
}
