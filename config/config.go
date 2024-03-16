package config

import (
	"berquerant/install-via-git-go/errorx"
	"errors"
	"io"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		URI      string            `yaml:"uri" json:"uri"`
		Branch   string            `yaml:"branch,omitempty" json:"branch,omitempty"`
		LocalDir string            `yaml:"locald,omitempty" json:"locald,omitempty"`
		LockFile string            `yaml:"lock,omitempty" json:"lock,omitempty"`
		Steps    Steps             `yaml:"steps,inline" json:"steps"`
		Env      map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
		Shell    []string          `yaml:"shell,omitempty" json:"shell,omitempty"`
	}

	Steps struct {
		Setup    []string `yaml:"setup,omitempty" json:"setup,omitempty"`
		Install  []string `yaml:"install,omitempty" json:"install,omitempty"`
		Rollback []string `yaml:"rollback,omitempty" json:"rollback,omitempty"`
		Skip     []string `yaml:"skip,omitempty" json:"skip,omitempty"`
		Check    []string `yaml:"check,omitempty" json:"check,omitempty"`
	}
)

func defaultConfig() Config {
	return Config{
		Branch:   "main",
		LockFile: "lock",
		LocalDir: "repo",
	}
}

var (
	ErrParse   = errors.New("Parse")
	ErrInvalid = errors.New("Invalid")
)

func Parse(r io.Reader) (*Config, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Join(ErrParse, err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, errors.Join(ErrParse, err)
	}

	if cfg.URI == "" {
		return nil, errorx.Errorf(ErrInvalid, "empty uri")
	}
	return &cfg, nil
}
