package config

import (
	"bufio"
	"context"
	"log"
	"os"
)

type StdInConfigurationFetcher struct {
}

func NewDebugConfigurationFetcher() IConfigurationFetcher {
	return &StdInConfigurationFetcher{}
}

func (*StdInConfigurationFetcher) GetString(_ context.Context, key string) (res string, err error) {
	scanner := bufio.NewScanner(os.Stdin)
	log.Printf("Please input the value of the Key:%s", key)
	scanner.Scan()
	return scanner.Text(), nil
}
