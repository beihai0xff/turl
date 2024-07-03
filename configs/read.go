package configs

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

// ReadFile reads a yaml file and returns a ServerConfig.
func ReadFile(path string, mp map[string]interface{}) (*ServerConfig, error) {
	// Load Yaml config.
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, err
	}

	if err := k.Load(confmap.Provider(mp, "."), nil); err != nil {
		return nil, err
	}

	_, _ = fmt.Fprint(os.Stdout, k.Sprint())

	var t ServerConfig
	if err := k.UnmarshalWithConf("", &t, koanf.UnmarshalConf{Tag: "json"}); err != nil {
		return nil, err
	}

	return &t, nil
}
