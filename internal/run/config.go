package run

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Hook struct {
	Command   string `yaml:"command"`
	TimeoutMs int    `yaml:"timeout_ms"`
}

type HooksConfig struct {
	OnWorktreeReady []Hook `yaml:"on_worktree_ready"`
}

type NotificationsConfig struct {
	OnComplete string `yaml:"on_complete"`
	OnError    string `yaml:"on_error"`
}

type Config struct {
	Hooks                   HooksConfig         `yaml:"hooks"`
	Notifications           NotificationsConfig `yaml:"notifications"`
	DefaultCompletionSignal string              `yaml:"default_completion_signal"`
}

func LoadConfig(repoRoot string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, ".amp.yaml"))
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	return cfg, yaml.Unmarshal(data, &cfg)
}
