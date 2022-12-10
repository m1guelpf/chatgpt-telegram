package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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
	for label, test := range map[string]struct {
		fileContent string
		envVars     map[string]string
		want        *EnvConfig
	}{
		"all values empty in file and env": {
			fileContent: `TELEGRAM_ID=
TELEGRAM_TOKEN=
EDIT_WAIT_SECONDS=`,
			want: &EnvConfig{
				TelegramID:      []int64{},
				TelegramToken:   "",
				EditWaitSeconds: 0,
			},
		},
		"all values provided in file, single TELEGRAM_ID": {
			fileContent: `TELEGRAM_ID=123
TELEGRAM_TOKEN=abc
EDIT_WAIT_SECONDS=10`,
			want: &EnvConfig{
				TelegramID:      []int64{123},
				TelegramToken:   "abc",
				EditWaitSeconds: 10,
			},
		},
		"multiple TELEGRAM_IDs provided in file": {
			fileContent: `TELEGRAM_ID=123,456
TELEGRAM_TOKEN=abc
EDIT_WAIT_SECONDS=10`,
			envVars: map[string]string{},
			want: &EnvConfig{
				TelegramID:      []int64{123, 456},
				TelegramToken:   "abc",
				EditWaitSeconds: 10,
			},
		},
		"env variables should override file values": {
			fileContent: `TELEGRAM_ID=123
TELEGRAM_TOKEN=abc
EDIT_WAIT_SECONDS=10`,
			envVars: map[string]string{
				"TELEGRAM_ID":       "456",
				"TELEGRAM_TOKEN":    "def",
				"EDIT_WAIT_SECONDS": "20",
			},
			want: &EnvConfig{
				TelegramID:      []int64{456},
				TelegramToken:   "def",
				EditWaitSeconds: 20,
			},
		},
		"multiple TELEGRAM_IDs provided in env": {
			fileContent: `TELEGRAM_ID=123
TELEGRAM_TOKEN=abc
EDIT_WAIT_SECONDS=10`,
			envVars: map[string]string{
				"TELEGRAM_ID": "456,789",
			},
			want: &EnvConfig{
				TelegramID:      []int64{456, 789},
				TelegramToken:   "abc",
				EditWaitSeconds: 10,
			},
		},
	} {
		t.Run(label, func(t *testing.T) {
			unset := setEnvVariables(test.envVars)
			t.Cleanup(unset)

			remove, err := createFile("test.env", test.fileContent)
			require.NoError(t, err)
			t.Cleanup(remove)

			cfg, err := LoadEnvConfig("test.env")
			require.NoError(t, err)
			require.Equal(t, test.want, cfg)
		})
	}

	//t.Run("all values empty in file", func(t *testing.T) {
	//
	//})
	//
	//t.Run("should load all values from file", func(t *testing.T) {
	//	cfg, err := LoadEnvConfig(fileName)
	//	require.NoError(t, err)
	//	require.Equal(t, []int64{123}, cfg.TelegramID)
	//	require.Equal(t, "abc", cfg.TelegramToken)
	//	require.Equal(t, 10, cfg.EditWaitSeconds)
	//})
	//
	//t.Run("env variables should override file values", func(t *testing.T) {
	//	unset := setEnvVariables(map[string]string{
	//		"TELEGRAM_ID":       "456,789",
	//		"TELEGRAM_TOKEN":    "def",
	//		"EDIT_WAIT_SECONDS": "20",
	//	})
	//
	//	cfg, err := LoadEnvConfig(fileName)
	//	unset()
	//	require.NoError(t, err)
	//	require.Equal(t, []int64{456, 789}, cfg.TelegramID)
	//	require.Equal(t, "def", cfg.TelegramToken)
	//	require.Equal(t, 20, cfg.EditWaitSeconds)
	//})
}
