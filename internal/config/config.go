package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Polymarket struct {
		APIKey     string `mapstructure:"api_key"`
		APISecret  string `mapstructure:"api_secret"`
		Passphrase string `mapstructure:"passphrase"`
	} `mapstructure:"polymarket"`
	Claude struct {
		APIKey   string `mapstructure:"api_key"`
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"claude"`
	Database struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"database"`
	UI struct {
		Theme string `mapstructure:"theme"`
	} `mapstructure:"ui"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Default values
	v.SetDefault("database.path", "polytracker.db")
	v.SetDefault("ui.theme", "dracula")
	v.SetDefault("claude.endpoint", "https://api.anthropic.com/v1/messages")

	// Environment variables
	v.SetEnvPrefix("POLYTRACKER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".polytracker"))
		}
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if not explicitly provided
			if configPath != "" {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func CreateDefaultConfig(path string) error {
	v := viper.New()
	v.Set("polymarket.api_key", "")
	v.Set("polymarket.api_secret", "")
	v.Set("polymarket.passphrase", "")
	v.Set("claude.api_key", "")
	v.Set("claude.endpoint", "https://api.anthropic.com/v1/messages")
	v.Set("database.path", "polytracker.db")
	v.Set("ui.theme", "dracula")

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return v.WriteConfigAs(path)
}
