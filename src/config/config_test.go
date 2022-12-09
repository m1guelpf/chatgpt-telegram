package config

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func createFile(name string, content string) (remove func(), err error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return nil, err
	}

	return func() {
		if err := os.Remove(name); err != nil {
			panic(fmt.Sprintf("failed to remove file: %s", err))
		}
	}, nil
}

func setEnvVariables(vals map[string]string) func() {
	for k, v := range vals {
		os.Setenv(k, v)
	}
	return func() {
		for k := range vals {
			os.Unsetenv(k)
		}
	}
}

func TestLoadEnvConfig(t *testing.T) {
	const fileName = "test.env"
	remove, err := createFile(fileName, `TELEGRAM_ID=123
TELEGRAM_TOKEN=abc
EDIT_WAIT_SECONDS=10`,
	)
	require.NoError(t, err)
	defer remove()

	t.Run("should load all values from file", func(t *testing.T) {
		cfg, err := LoadEnvConfig(fileName)
		require.NoError(t, err)
		require.Equal(t, []int64{123}, cfg.TelegramID)
		require.Equal(t, "abc", cfg.TelegramToken)
		require.Equal(t, 10, cfg.EditWaitSeconds)
	})

	t.Run("env variables should override file values", func(t *testing.T) {
		unset := setEnvVariables(map[string]string{
			"TELEGRAM_ID":       "456,789",
			"TELEGRAM_TOKEN":    "def",
			"EDIT_WAIT_SECONDS": "20",
		})

		cfg, err := LoadEnvConfig(fileName)
		unset()
		require.NoError(t, err)
		require.Equal(t, []int64{456, 789}, cfg.TelegramID)
		require.Equal(t, "def", cfg.TelegramToken)
		require.Equal(t, 20, cfg.EditWaitSeconds)
	})
}
