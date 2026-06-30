package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all user-configurable application settings.
type Config struct {
	// Hosts is the list of upstream API mirrors tried in parallel.
	Hosts []string `yaml:"hosts" json:"hosts"`
	// Timeout is the HTTP request timeout in seconds.
	Timeout int `yaml:"timeout" json:"timeout"`
	// CacheTTLs controls how long search and content results are cached.
	CacheTTLs CacheConfig `yaml:"cache_ttls" json:"cache_ttls"`
	// LogLevel controls the minimum slog level: debug, info, warn, or error.
	LogLevel string `yaml:"log_level" json:"log_level"`
	// LogPath is the full path to the log file. An empty value means stderr only.
	LogPath string `yaml:"log_path" json:"log_path"`
}

// CacheConfig holds TTL values in seconds for different cache types.
type CacheConfig struct {
	Search  int `yaml:"search" json:"search"`
	Content int `yaml:"content" json:"content"`
}

// Default returns a Config with sensible built-in defaults that keep the
// application usable without any config file on disk.
func Default() *Config {
	return &Config{
		Hosts: []string{
			"https://v1.gyks.cf",
			"https://v2.gyks.cf",
			"https://v3.gyks.cf",
			"https://v4.gyks.cf",
			"https://v5.gyks.cf",
			"https://v6.gyks.cf",
			"https://v7.gyks.cf",
		},
		Timeout:  15,
		LogLevel: "info",
		LogPath:  defaultLogPath(),
	}
}

// defaultLogPath builds a path under the OS config directory so we never
// hardcode /tmp. Falls back to the OS temp directory if UserConfigDir fails.
func defaultLogPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "yueduqi.log")
	}
	return filepath.Join(dir, "yueduqi", "yueduqi.log")
}

// Load reads config.yaml from the OS config directory (~/.config/yueduqi/
// on Linux). When the file is missing Load returns Default() with a nil
// error so callers don't need to special-case "not found".
func Load() (*Config, error) {
	cfg := Default()

	dir, err := os.UserConfigDir()
	if err != nil {
		return cfg, fmt.Errorf("cannot determine config dir: %w", err)
	}

	configPath := filepath.Join(dir, "yueduqi", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config file is not an error
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// ConfigDir returns the yueduqi config directory, creating it if needed.
func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "yueduqi")
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}
