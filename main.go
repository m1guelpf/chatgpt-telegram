package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)

func main() {
	runOptions := playwright.RunOptions{
		Browsers: []string{"chromium"},
		Verbose:  false,
	}
	err := playwright.Install(&runOptions)
	if err != nil {
		log.Fatalf("Couldn't install headless browser: %v", err)
	}

	pw, err := playwright.Run(&runOptions)
	if err != nil {
		log.Fatalf("Couldn't start headless browser: %v", err)
	}

	browser, page := launchBrowser(pw, "https://chat.openai.com", true)

	for !isLoggedIn(page) {
		cookies := <-logIn(pw)
		for _, cookie := range cookies {
			convertedCookie := playwright.BrowserContextAddCookiesOptionsCookies{
				Name:     &cookie.Name,
				Value:    &cookie.Value,
				Domain:   &cookie.Domain,
				Path:     &cookie.Path,
				Expires:  &cookie.Expires,
				Secure:   &cookie.Secure,
				HttpOnly: &cookie.HttpOnly,
				SameSite: &cookie.SameSite,
			}
			if err := browser.AddCookies(convertedCookie); err != nil {
				log.Fatalf("Couldn't add cookies: %v", err)
			}
		}

		if _, err = page.Goto("https://chat.openai.com"); err != nil {
			log.Fatalf("Couldn't reload website: %v", err)
		}
	}

	if _, err := page.Evaluate("localStorage.setItem('oai/apps/hasSeenOnboarding/chat', 'true')"); err != nil {
		log.Fatalf("Couldn't update localstorage: %v", err)
	}
	if _, err = page.Reload(); err != nil {
		log.Fatalf("Couldn't reload website: %v", err)
	}
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
		if err = browser.Close(); err != nil {
			log.Fatalf("could not close browser: %v", err)
		}
		if err = pw.Stop(); err != nil {
			log.Fatalf("could not stop Playwright: %v", err)
		}
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

		if !update.Message.IsCommand() {
			response, err := query_chatgpt(update.Message.Text, page, bot, tgbotapi.NewChatAction(update.Message.Chat.ID, "typing"))
			if err != nil {
				msg.Text = fmt.Sprintf("Error: %v", err)
			} else {
				msg.Text = response
			}

			if _, err := bot.Send(msg); err != nil {
				log.Fatalf("Couldn't send message: %v", err)
			}
			continue
		}

		switch update.Message.Command() {
		case "help":
			msg.Text = "Help!"
		case "start":
			msg.Text = "Hi :)"
		case "reload":
			if _, err = page.Reload(); err != nil {
				log.Fatalf("Couldn't reload website: %v", err)
			}
			msg.Text = "New conversation!"
		default:
			continue
		}

		if _, err := bot.Send(msg); err != nil {
			log.Fatalf("Couldn't send message: %v", err)
		}
	}
}

func query_chatgpt(query string, page playwright.Page, bot *tgbotapi.BotAPI, showTyping tgbotapi.ChatActionConfig) (string, error) {
	input := getChatBox(page)

	if err := input.Click(); err != nil {
		log.Fatalf("Couldn't type query: %v", err)
	}
	if err := input.Fill(query); err != nil {
		log.Fatalf("Couldn't type query: %v", err)
	}
	if err := input.Press("Enter"); err != nil {
		log.Fatalf("Couldn't type query: %v", err)
	}

	loading, err := page.QuerySelectorAll(
		"button[class^='PromptTextarea__PositionSubmit']>.text-2xl",
	)
	if err != nil {
		log.Fatalf("Couldn't get loading element: %v", err)
	}
	bot.Request(showTyping)
	start_time := time.Now()

	for len(loading) > 0 {
		if (time.Now().Sub(start_time)).Seconds() > 45 {
			return "", errors.New("Loading took too long")
		}

		time.Sleep(500 * time.Millisecond)

		loading, err = page.QuerySelectorAll(
			"button[class^='PromptTextarea__PositionSubmit']>.text-2xl",
		)
		if err != nil {
			return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
		}
		bot.Request(showTyping)
	}

	pageElements, err := page.QuerySelectorAll("div[class*='ConversationItem__Message']")
	if err != nil {
		return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
	}
	lastElement := pageElements[len(pageElements)-1]
	prose, err := lastElement.QuerySelector(".prose")
	if err != nil {
		return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
	}
	codeBlocks, err := prose.QuerySelectorAll("pre")
	if err != nil {
		return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
	}

	response := ""
	if len(codeBlocks) > 0 {
		children, err := prose.QuerySelectorAll("p, pre")
		if err != nil {
			return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
		}

		for _, child := range children {
			tagName, err := child.GetProperty("tagName")
			if err != nil {
				return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
			}
			if tagName.String() == "PRE" {
				codeContainer, err := child.QuerySelector("code")
				if err != nil {
					return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
				}
				innerText, err := codeContainer.InnerText()
				if err != nil {
					return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
				}
				response += fmt.Sprintf("\n\n```\n%s\n```", escapeMarkdown(innerText))
			} else {
				text, err := child.InnerText()
				if err != nil {
					return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
				}
				response += escapeMarkdown(text)
			}
		}
		response = strings.ReplaceAll(response, "<code>", "`")
		response = strings.ReplaceAll(response, "</code>", "`")
	} else {
		innerText, err := prose.InnerText()
		if err != nil {
			return "", errors.New("Something went wrong! Try running /reload, or restart the CLI.")
		}
		response = escapeMarkdown(innerText)
	}

	return response, nil
}

func launchBrowser(pw *playwright.Playwright, url string, headless bool) (playwright.BrowserContext, playwright.Page) {
	browser, err := pw.Chromium.LaunchPersistentContext("/tmp/chatgpt", playwright.BrowserTypeLaunchPersistentContextOptions{Headless: playwright.Bool(headless)})
	if err != nil {
		log.Fatalf("Couldn't launch headless browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Couldn't create a new tab on headless browser: %v", err)
	}

	if _, err = page.Goto(url); err != nil {
		log.Fatalf("Couldn't open website: %v", err)
	}

	return browser, page
}

func isLoggedIn(page playwright.Page) bool {
	return page.URL() == "https://chat.openai.com/chat"
}

func getChatBox(page playwright.Page) playwright.ElementHandle {
	input, err := page.QuerySelector("textarea")
	if err != nil {
		log.Fatalf("Couldn't get chatbox: %v", err)
	}

	return input
}

func logIn(pw *playwright.Playwright) <-chan []*playwright.BrowserContextCookiesResult {
	var lock sync.Mutex
	r := make(chan []*playwright.BrowserContextCookiesResult)

	lock.Lock()
	go func() {
		defer close(r)
		defer lock.Unlock()

		browser, page := launchBrowser(pw, "https://chat.openai.com/", false)
		log.Println("Please log in to OpenAI Chat")

		page.On("framenavigated", func(frame playwright.Frame) {
			if frame.URL() != "https://chat.openai.com/chat" {
				return
			}

			lock.Unlock()
		})

		lock.Lock()

		cookies, err := browser.Cookies("https://chat.openai.com/")
		if err != nil {
			log.Fatalf("Couldn't store authentication: %v", err)
		}

		if err := browser.Close(); err != nil {
			log.Fatalf("could not close browser: %v", err)
		}

		r <- cookies
	}()

	return r
}

func escapeMarkdown(text string) string {
	escape_chars := []string{"_", "*", "`", "["}

	for _, char := range escape_chars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}

	return text
}
