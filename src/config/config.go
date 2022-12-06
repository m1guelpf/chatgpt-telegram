package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	OpenAISession string
}

// init tries to read the config from the file, and creates it if it doesn't exist.
func Init() (Config, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return Config{}, errors.New(fmt.Sprintf("Couldn't get user config dir: %v", err))
	}
	viper.SetConfigType("json")
	viper.SetConfigName("chatgpt")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				return Config{}, errors.New(fmt.Sprintf("Couldn't create config file: %v", err))
			}
		} else {
			return Config{}, errors.New(fmt.Sprintf("Couldn't read config file: %v", err))
		}
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return Config{}, errors.New(fmt.Sprintf("Error parsing config: %v", err))
	}

	return cfg, nil
}

// key should be part of the Config struct
func (cfg *Config) Set(key string, value interface{}) error {
	viper.Set(key, value)

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing config: %v", err))
	}

	return viper.WriteConfig()
}
