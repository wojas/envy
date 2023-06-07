package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/wojas/envy/paths"
	"gopkg.in/yaml.v2"
)

type Config struct {
	// Dirs under which we will load .envy files
	TrustedPaths []string `yaml:"trusted_paths"`
	// Always load ~/.envy and check for dirs there, even if outside of the home dir
	AlwaysLoadHome bool `yaml:"always_load_home"`
	// Allows changing tab and other colors with ENVY_COLOR
	Colors map[string]string `yaml:"colors"`
}

// Check validates a Config instance
func (c Config) Check() error {
	if len(c.TrustedPaths) == 0 {
		return fmt.Errorf("trusted_paths (list) is empty")
	}
	return nil
}

// String returns the config as a YAML string
func (c Config) String() string {
	y, err := yaml.Marshal(c)
	if err != nil {
		log.Panicf("YAML marshal of config failed: %v", err) // Should never happen
	}
	return string(y)
}

// LoadYAML loads config from YAML. Any set value overwrites any existing value,
// but omitted keys are untouched.
func (c *Config) LoadYAML(yamlContents []byte) error {
	return yaml.Unmarshal(yamlContents, c)
}

// LoadYAMLFile loads config from a YAML file
func (c *Config) LoadYAMLFile(fpath string) error {
	contents, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	return c.LoadYAML(contents)
}

// Default returns a Config with default settings
func Default() *Config {
	var trusted []string
	if home, err := paths.HomeDir(); err == nil {
		trusted = append(trusted, home)
	}
	return &Config{
		TrustedPaths:   trusted,
		AlwaysLoadHome: true,
	}
}
