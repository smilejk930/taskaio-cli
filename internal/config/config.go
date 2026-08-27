package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBaseURL = "http://localhost:3000"
	DefaultOutput  = "json"
	DefaultTimeout = 30 * time.Second
)

// Config represents taskaio CLI configuration.
type Config struct {
	BaseURL    string        `yaml:"base_url"`
	Token      string        `yaml:"token"`
	Output     string        `yaml:"output"`
	TimeoutRaw string        `yaml:"timeout,omitempty"`
	Timeout    time.Duration `yaml:"-"`
	ConfigPath string        `yaml:"-"`
}

// DefaultConfigPath returns ~/.config/taskaio/config.yaml or $XDG_CONFIG_HOME/taskaio/config.yaml
func DefaultConfigPath() (string, error) {
	if custom := os.Getenv("TASKAIO_CONFIG"); custom != "" {
		return custom, nil
	}

	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, "taskaio", "config.yaml"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user home directory: %w", err)
	}
	return filepath.Join(home, ".config", "taskaio", "config.yaml"), nil
}

// LoadConfig loads configuration according to precedence: Flags > Env > File > Defaults.
func LoadConfig(flagConfigPath, flagBaseURL, flagToken, flagOutput string, flagTimeout time.Duration) (*Config, error) {
	cfg := &Config{
		BaseURL: DefaultBaseURL,
		Output:  DefaultOutput,
		Timeout: DefaultTimeout,
	}

	// 1. Determine config path
	configPath := flagConfigPath
	if configPath == "" {
		var err error
		configPath, err = DefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}
	cfg.ConfigPath = configPath

	// 2. Load from YAML file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file at %s: %w", configPath, err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse yaml config at %s: %w", configPath, err)
		}

		if cfg.TimeoutRaw != "" {
			d, err := time.ParseDuration(cfg.TimeoutRaw)
			if err != nil || d <= 0 {
				return nil, fmt.Errorf("invalid timeout in config file %s: %q", configPath, cfg.TimeoutRaw)
			}
			cfg.Timeout = d
		}
	}

	// 3. Override with Environment Variables (TASKAIO_*)
	if envURL := os.Getenv("TASKAIO_BASE_URL"); envURL != "" {
		cfg.BaseURL = envURL
	}
	if envToken := os.Getenv("TASKAIO_TOKEN"); envToken != "" {
		cfg.Token = envToken
	}
	if envOutput := os.Getenv("TASKAIO_OUTPUT"); envOutput != "" {
		cfg.Output = envOutput
	}
	if envTimeout := os.Getenv("TASKAIO_TIMEOUT"); envTimeout != "" {
		d, err := time.ParseDuration(envTimeout)
		if err != nil || d <= 0 {
			return nil, fmt.Errorf("invalid TASKAIO_TIMEOUT: %q", envTimeout)
		}
		cfg.Timeout = d
	}

	// 4. Override with CLI Flags
	if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}
	if flagToken != "" {
		cfg.Token = flagToken
	}
	if flagOutput != "" {
		cfg.Output = flagOutput
	}
	if flagTimeout > 0 {
		cfg.Timeout = flagTimeout
	}

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Output != "json" && cfg.Output != "table" {
		return nil, fmt.Errorf("invalid output format %q: expected json or table", cfg.Output)
	}
	return cfg, nil
}

// SaveConfig writes the configuration to disk with strict 0600 file mode.
func SaveConfig(cfg *Config, targetPath string) error {
	if targetPath == "" {
		var err error
		targetPath, err = DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if cfg.Timeout > 0 {
		cfg.TimeoutRaw = cfg.Timeout.String()
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file with strict 0600 permissions
	if err := os.WriteFile(targetPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Ensure permissions are strictly 0600 even if file already existed with different perms
	if err := os.Chmod(targetPath, 0600); err != nil {
		return fmt.Errorf("failed to chmod config file to 0600: %w", err)
	}

	return nil
}
