package configs

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

// ReadFile reads a yaml file and returns a ServerConfig.
func ReadFile(path string) (*ServerConfig, error) {
	// Load Yaml config.
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, err
	}

	fmt.Println(k.Sprint())

	var t ServerConfig
	if err := k.UnmarshalWithConf("", &t, koanf.UnmarshalConf{Tag: "json"}); err != nil {
		return nil, err
	}

	return &t, nil
}
