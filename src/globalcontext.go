package src

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shamu00/chatgpt-telegram/src/args"
	"github.com/shamu00/chatgpt-telegram/src/chatgpt"
	libconfig "github.com/shamu00/chatgpt-telegram/src/config"
	"log"
	"os"
)

const (
	defaultHttpEndPoint = "https://config-center.azconfig.io"
)

type GlobalContext struct {
	Context       context.Context
	TelegramId    string
	TelegramToken string
	ChatClient    chatgpt.IChatClient
	Bot           *tgbotapi.BotAPI
}

func PrepareContext() GlobalContext {
	var result = GlobalContext{
		Context: context.Background(),
	}
	var arg = args.Parse(os.Args)
	var configurationFetcher libconfig.IConfigurationFetcher = libconfig.NewDebugConfigurationFetcher()
	if !arg.DebugMode {
		azureCredential := os.Getenv(libconfig.AzureConfigCenterCredential)
		azureSecret := os.Getenv(libconfig.AzureConfigCenterSecret)
		// TODO debug delete
		azureCredential = "DiYj-lb-s0:yNmHMgvP1QNp4dq8Lqv5"
		azureSecret = "Q1QJSgYk4P1fP1B/Eufj+C5RHJUvLKwTAL+D2Pgqw7k="
		configurationFetcher = libconfig.NewAzureConfigurationFetcher(defaultHttpEndPoint, azureCredential, azureSecret)
	}
	tgId, err := configurationFetcher.GetString(result.Context, libconfig.KeyTelegramId)
	if err != nil {
		log.Fatalf("Couldn't fetch KeyTelegramId, err:%v", err)
	}
	result.TelegramId = tgId
	tgToken, err := configurationFetcher.GetString(result.Context, libconfig.KeyTelegramToken)
	if err != nil {
		log.Fatalf("Couldn't fetch KeyTelegramToken, err:%v", err)
	}
	result.TelegramToken = tgToken
	chatApiKey, err := configurationFetcher.GetString(result.Context, libconfig.KeyChatGptOpenKey)
	if err != nil {
		log.Fatalf("Couldn't fetch KeyChatOpenApiKey, err:%v", err)
	}
	result.ChatClient = chatgpt.NewChatGptClient(chatApiKey)
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Fatalf("Couldn't start Telegram bot, err:%v", err)
	}
	result.Bot = bot
	return result
}
