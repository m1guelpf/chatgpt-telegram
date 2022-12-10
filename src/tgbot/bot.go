package tgbot

import (
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m1guelpf/chatgpt-telegram/src/chatgpt"
	"github.com/m1guelpf/chatgpt-telegram/src/markdown"
	"github.com/m1guelpf/chatgpt-telegram/src/ratelimit"
)

type Bot struct {
	Username     string
	api          *tgbotapi.BotAPI
	editInterval time.Duration
}

func New(token string, editInterval time.Duration) (*Bot, error) {
	var api *tgbotapi.BotAPI
	var err error
	apiEndpoint, exist := os.LookupEnv("TELEGRAM_API_ENDPOINT")
	if exist && apiEndpoint != "" {
		api, err = tgbotapi.NewBotAPIWithAPIEndpoint(token, apiEndpoint)
	} else {
		api, err = tgbotapi.NewBotAPI(token)
	}
	if err != nil {
		return nil, err
	}

	return &Bot{
		Username:     api.Self.UserName,
		api:          api,
		editInterval: editInterval,
	}, nil
}

func (b *Bot) GetUpdatesChan() tgbotapi.UpdatesChannel {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 30
	return b.api.GetUpdatesChan(cfg)
}

func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *Bot) Send(chatID int64, replyTo int, text string) (tgbotapi.Message, error) {
	text = markdown.EnsureFormatting(text)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyTo
	return b.api.Send(msg)
}

func (b *Bot) SendEdit(chatID int64, messageID int, text string) error {
	text = markdown.EnsureFormatting(text)
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "Markdown"
	if _, err := b.api.Send(msg); err != nil {
		if err.Error() == "Bad Request: message is not modified: specified new message content and reply markup are exactly the same as a current content and reply markup of the message" {
			return nil
		}
		return err
	}
	return nil
}

func (b *Bot) SendTyping(chatID int64) {
	if _, err := b.api.Request(tgbotapi.NewChatAction(chatID, "typing")); err != nil {
		log.Printf("Couldn't send typing action: %v", err)
	}
}

func (b *Bot) SendAsLiveOutput(chatID int64, replyTo int, feed chan chatgpt.ChatResponse) {
	debouncedType := ratelimit.Debounce(10*time.Second, func() { b.SendTyping(chatID) })
	debouncedEdit := ratelimit.DebounceWithArgs(b.editInterval, func(text interface{}, messageId interface{}) {
		if err := b.SendEdit(chatID, messageId.(int), text.(string)); err != nil {
			log.Printf("Couldn't edit message: %v", err)
		}
	})

	var message tgbotapi.Message
	var lastResp string

pollResponse:
	for {
		debouncedType()

		select {
		case response, ok := <-feed:
			if !ok {
				break pollResponse
			}

			lastResp = response.Message

			if message.MessageID == 0 {
				var err error
				if message, err = b.Send(chatID, replyTo, lastResp); err != nil {
					log.Fatalf("Couldn't send message: %v", err)
				}
			} else {
				debouncedEdit(lastResp, message.MessageID)
			}
		}
	}

	if err := b.SendEdit(chatID, message.MessageID, lastResp); err != nil {
		log.Printf("Couldn't perform final edit on message: %v", err)
	}
}
