package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/m1guelpf/chatgpt-telegram/src/chatgpt"
	"github.com/m1guelpf/chatgpt-telegram/src/config"
	"github.com/m1guelpf/chatgpt-telegram/src/markdown"
	"github.com/m1guelpf/chatgpt-telegram/src/ratelimit"
	"github.com/m1guelpf/chatgpt-telegram/src/session"
)

type Conversation struct {
	ConversationID string
	LastMessageID  string
}

func main() {
	config, err := config.Init()
	if err != nil {
		log.Fatalf("Couldn't load config: %v", err)
	}

	if config.OpenAISession == "" {
		session, err := session.GetSession()
		if err != nil {
			log.Fatalf("Couldn't get OpenAI session: %v", err)
		}

		err = config.Set("OpenAISession", session)
		if err != nil {
			log.Fatalf("Couldn't save OpenAI session: %v", err)
		}
	}

	chatGPT := chatgpt.Init(config)
	log.Println("Started ChatGPT")

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Couldn't load .env file: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatalf("Couldn't start Telegram bot: %v", err)
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		bot.StopReceivingUpdates()
		os.Exit(0)
	}()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	log.Printf("Started Telegram bot! Message @%s to start.", bot.Self.UserName)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ParseMode = "Markdown"

		userId := strconv.FormatInt(update.Message.Chat.ID, 10)
		if os.Getenv("TELEGRAM_ID") != "" && userId != os.Getenv("TELEGRAM_ID") {
			msg.Text = "You are not authorized to use this bot."
			bot.Send(msg)
			continue
		}

		bot.Request(tgbotapi.NewChatAction(update.Message.Chat.ID, "typing"))
		if !update.Message.IsCommand() {
			feed, err := chatGPT.SendMessage(update.Message.Text, update.Message.Chat.ID)
			if err != nil {
				msg.Text = fmt.Sprintf("Error: %v", err)
			}

			var message tgbotapi.Message
			var lastResp string

			debouncedType := ratelimit.Debounce((10 * time.Second), func() {
				bot.Request(tgbotapi.NewChatAction(update.Message.Chat.ID, "typing"))
			})
			debouncedEdit := ratelimit.DebounceWithArgs((1 * time.Second), func(text interface{}, messageId interface{}) {
				_, err = bot.Request(tgbotapi.EditMessageTextConfig{
					BaseEdit: tgbotapi.BaseEdit{
						ChatID:    msg.ChatID,
						MessageID: messageId.(int),
					},
					Text:      text.(string),
					ParseMode: "Markdown",
				})

				if err != nil {
					if err.Error() == "Bad Request: message is not modified: specified new message content and reply markup are exactly the same as a current content and reply markup of the message" {
						return
					}

					log.Printf("Couldn't edit message: %v", err)
				}
			})

		pollResponse:
			for {
				debouncedType()

				select {
				case response, ok := <-feed:
					if !ok {
						break pollResponse
					}

					lastResp = markdown.EnsureFormatting(response.Message)
					msg.Text = lastResp

					if message.MessageID == 0 {
						message, err = bot.Send(msg)
						if err != nil {
							log.Fatalf("Couldn't send message: %v", err)
						}
					} else {
						debouncedEdit(lastResp, message.MessageID)
					}
				}
			}

			_, err = bot.Request(tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:    msg.ChatID,
					MessageID: message.MessageID,
				},
				Text:      lastResp,
				ParseMode: "Markdown",
			})

			if err != nil {
				if err.Error() == "Bad Request: message is not modified: specified new message content and reply markup are exactly the same as a current content and reply markup of the message" {
					continue
				}

				log.Printf("Couldn't perform final edit on message: %v", err)
			}

			continue
		}

		switch update.Message.Command() {
		case "help":
			msg.Text = "Send a message to start talking with ChatGPT. You can use /reload at any point to clear the conversation history and start from scratch (don't worry, it won't delete the Telegram messages)."
		case "start":
			msg.Text = "Send a message to start talking with ChatGPT. You can use /reload at any point to clear the conversation history and start from scratch (don't worry, it won't delete the Telegram messages)."
		case "reload":
			chatGPT.ResetConversation(update.Message.Chat.ID)
			msg.Text = "Started a new conversation. Enjoy!"
		default:
			continue
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Couldn't send message: %v", err)
			continue
		}
	}
}
