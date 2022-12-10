package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gogpt "github.com/sashabaranov/go-gpt3"
	"github.com/shamu00/chatgpt-telegram/src"
	"github.com/shamu00/chatgpt-telegram/src/markdown"
	"github.com/shamu00/chatgpt-telegram/src/util"
	"log"
	"strconv"
	"time"
)

type Conversation struct {
	ConversationID string
	LastMessageID  string
}

func HandleBotMessage(ctx src.GlobalContext, updates tgbotapi.UpdatesChannel) {
	bot := ctx.Bot
	userConversations := make(map[int64]Conversation)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[Input]bot says:%v\n", update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "linlin says:")
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ParseMode = tgbotapi.ModeMarkdown

		//userId := strconv.FormatInt(update.Message.Chat.ID, 10)
		//if ctx.TelegramId != "" && userId != ctx.TelegramId {
		//	msg.Text = fmt.Sprintf("You are not authorized to use this bot. userID:%v", userId)
		//	ctx.Bot.Send(msg)
		//	continue
		//}
		ctx.Bot.Request(tgbotapi.NewChatAction(update.Message.Chat.ID, "typing"))
		if update.Message.IsCommand() {
			handleCommand(update.Message.Command(), &msg, userConversations, update.Message.Chat.ID)
			sendMsg(bot, msg)
		}
		// message type
		handleChat(ctx, update, &msg)
		sendMsg(bot, msg)
	}

}

func handleChat(ctx src.GlobalContext,
	update tgbotapi.Update,
	msg *tgbotapi.MessageConfig,
) {
	creq := &gogpt.CompletionRequest{
		Prompt:    update.Message.Text,
		Suffix:    "",
		LogProbs:  0,
		Stop:      nil,
		LogitBias: nil,
		User:      strconv.FormatInt(update.Message.From.ID, 10),
	}
	msgCh := make(chan *gogpt.CompletionResponse)
	go func() {
		result := &gogpt.CompletionResponse{}
		err := util.Retry(3, 500*time.Millisecond, func() error {
			var e error
			result, e = ctx.ChatClient.Talk(ctx.Context, creq)
			return e
		})
		if err != nil {
			result = &gogpt.CompletionResponse{Choices: []gogpt.CompletionChoice{{Text: err.Error()}}}
			log.Printf("[Error] got error when calling gogpt:%v\n", err)
		} else if len(result.Choices) == 0 {
			result.Choices = []gogpt.CompletionChoice{{Text: "no response"}}
			log.Printf("[Warn] got no choice when calling gogpt:%v\n", err)
		}
		msgCh <- result
		return
	}()
loop:
	for {
		ticker := time.NewTicker(5 * time.Second)
		select {
		case <-ticker.C:
			ctx.Bot.Request(tgbotapi.NewChatAction(update.Message.Chat.ID, "typing"))
		case result := <-msgCh:
			talk := result.Choices[0].Text
			mdTalk := markdown.EnsureFormatting(talk)
			msg.Text = mdTalk
			break loop
		}
	}
	return
}

func handleCommand(
	command string,
	msg *tgbotapi.MessageConfig,
	userConversations map[int64]Conversation,
	userId int64) {
	switch command {
	case "help":
		msg.Text = "Send a message to start talking with ChatGPT. You can use /reload at any point to clear the conversation history and start from scratch (don't worry, it won't delete the Telegram messages)."
	case "start":
		msg.Text = "Send a message to start talking with ChatGPT. You can use /reload at any point to clear the conversation history and start from scratch (don't worry, it won't delete the Telegram messages)."
	case "reload":
		userConversations[userId] = Conversation{}
		msg.Text = "Started a new conversation. Enjoy!"
	default:

	}
	return
}

func sendMsg(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	res, err := bot.Send(msg)
	if err != nil {
		log.Printf("[ERROR]Couldn't send message: %v", err)
		return err
	}
	log.Printf("[DEBUG]Message:%+v", res)
	return nil
}
