package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type EnvConfig struct {
	TelegramID      []int64 `mapstructure:"TELEGRAM_ID"`
	TelegramToken   string  `mapstructure:"TELEGRAM_TOKEN"`
	EditWaitSeconds int     `mapstructure:"EDIT_WAIT_SECONDS"`
}

func (e *EnvConfig) HasTelegramID(id int64) bool {
	for _, v := range e.TelegramID {
		if v == id {
			return true
		}
	}
	return false
}

// LoadEnvConfig loads config from .env file, variables from environment take precedence if provided.
func LoadEnvConfig(path string) (*EnvConfig, error) {
	if !fileExists(path) {
		return nil, fmt.Errorf("config file '%s' does not exist", path)
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg EnvConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return os.IsExist(err)
	}
	return true
}

func (e *EnvConfig) Validate() error {
	if e.TelegramToken == "" {
		return errors.New("TELEGRAM_TOKEN is not set")
	}
	if len(e.TelegramID) == 0 {
		log.Printf("TELEGRAM_ID is not set, all users will be able to use the bot")
	}
	if e.EditWaitSeconds < 0 {
		log.Printf("EDIT_WAIT_SECONDS not set, defaulting to 1")
		e.EditWaitSeconds = 1
	}
	return nil
}
