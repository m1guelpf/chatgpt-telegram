package main

import (
	"github.com/shamu00/chatgpt-telegram/src"
	"github.com/shamu00/chatgpt-telegram/src/tgbot"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	libconfig "github.com/shamu00/chatgpt-telegram/src/config"
)

func mustInit() {
	rand.Seed(time.Now().UnixNano())
	libconfig.InitConfigurationFetcher()
	return
}

func main() {
	mustInit()
	ctx := src.PrepareContext()
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ctx.Bot.StopReceivingUpdates()
		os.Exit(0)
	}()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := ctx.Bot.GetUpdatesChan(updateConfig)

	log.Println("start handling messages...")
	tgbot.HandleBotMessage(ctx, updates)

}
