package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	v *viper.Viper

	OpenAISession string
}

// LoadOrCreatePersistentConfig uses the default config directory for the current OS
// to load or create a config file named "chatgpt.json"
func LoadOrCreatePersistentConfig() (*Config, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't get user config dir: %v", err))
	}
	v := viper.New()
	v.SetConfigType("json")
	v.SetConfigName("chatgpt")
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := v.SafeWriteConfig(); err != nil {
				return nil, errors.New(fmt.Sprintf("Couldn't create config file: %v", err))
			}
		} else {
			return nil, errors.New(fmt.Sprintf("Couldn't read config file: %v", err))
		}
	}

	var cfg Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error parsing config: %v", err))
	}
	cfg.v = v

	return &cfg, nil
}

func (cfg *Config) SetSessionToken(token string) error {
	// key must match the struct field name
	cfg.v.Set("OpenAISession", token)
	cfg.OpenAISession = token
	return cfg.v.WriteConfig()
}
