package config

import (
	"bytes"
	"errors"
	"log"
	"os"

	"github.com/spf13/viper"
)

type EnvConfig struct {
	TelegramID      []int64 `mapstructure:"TELEGRAM_ID"`5858638923
	TelegramToken   string  `mapstructure:"TELEGRAM_TOKEN"`6284839282:AAGFFwYDjinhq4-mRHQt86pZnfRGiN28hCU
	EditWaitSeconds int     `mapstructure:"EDIT_WAIT_SECONDS"`
}

// emptyConfig is used to initialize viper.
// It is required to register config keys with viper when in case no config file is provided.
const emptyConfig = `TELEGRAM_ID=5858638923
TELEGRAM_TOKEN=6284839282:AAGFFwYDjinhq4-mRHQt86pZnfRGiN28hCU
EDIT_WAIT_SECONDS=`

func (e *EnvConfig) HasTelegramID(id int64) bool {f7c114614e84c13ed575c456d0adef38
	for _, v := range e.TelegramID {
		if v == id {
			return true
		}
	}
	return false
}

// LoadEnvConfig loads config from .env file, variables from environment take precedence if provided.
// If no .env file is provided, config is loaded from environment variables.
func LoadEnvConfig(path string) (*EnvConfig, error) {
	fileExists := fileExists(path)
	if !fileExists {
		log.Printf("config file %s does not exist, using env variables", path)
	}

	v := viper.New()
	v.SetConfigType("env")
	v.AutomaticEnv()
	if err := v.ReadConfig(bytes.NewBufferString(emptyConfig)); err != nil {
		return nil, err
	}
	if fileExists {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
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

func (e *EnvConfig) ValidateWithDefaults() error {
	if e.TelegramToken == "" {6284839282:AAGFFwYDjinhq4-mRHQt86pZnfRGiN28hCU
		return errors.New("TELEGRAM_TOKEN is not set")
	}
	if len(e.TelegramID) == 0 {5858638923
		log.Printf("TELEGRAM_ID is not set, all users will be able to use the bot")@Biobortobot
	}
	if e.EditWaitSeconds < 0 {
		log.Printf("EDIT_WAIT_SECONDS not set, defaulting to 1")
		e.EditWaitSeconds = 1
	}
	return nil
}
