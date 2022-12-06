package session

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/m1guelpf/chatgpt-telegram/src/ref"
	"github.com/playwright-community/playwright-go"
)

func GetSession() (string, error) {
	runOptions := playwright.RunOptions{
		Browsers: []string{"chromium"},
		Verbose:  false,
	}
	err := playwright.Install(&runOptions)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't install headless browser: %v", err))
	}

	pw, err := playwright.Run(&runOptions)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't start headless browser: %v", err))
	}

	browser, page, err := launchBrowser(pw, "https://chat.openai.com", true)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't launch headless browser: %v", err))
	}

	for page.URL() != "https://chat.openai.com/chat" {
		result := <-logIn(pw)
		if result.Error != nil {
			return "", errors.New(fmt.Sprintf("Couldn't log in: %v", result.Error))
		}

		authCookie := playwright.BrowserContextAddCookiesOptionsCookies{
			Path:     ref.Of("/"),
			Secure:   ref.Of(true),
			HttpOnly: ref.Of(true),
			Value:    ref.Of(result.SessionToken),
			Domain:   ref.Of("chat.openai.com"),
			SameSite: playwright.SameSiteAttributeLax,
			Name:     ref.Of("__Secure-next-auth.session-token"),
			Expires:  ref.Of(float64(time.Now().AddDate(0, 1, 0).Unix())),
		}

		if err := browser.AddCookies(authCookie); err != nil {
			return "", errors.New(fmt.Sprintf("Couldn't save session to browser: %v", err))
		}

		if _, err = page.Goto("https://chat.openai.com/chat"); err != nil {
			return "", errors.New(fmt.Sprintf("Couldn't reload page: %v", err))
		}
	}

	sessionToken, err := getSessionCookie(browser)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't get session token: %v", err))
	}

	if err := browser.Close(); err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't close headless browser: %v", err))
	}
	if err := pw.Stop(); err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't stop headless browser: %v", err))
	}

	return sessionToken, nil
}

func launchBrowser(pw *playwright.Playwright, url string, headless bool) (playwright.BrowserContext, playwright.Page, error) {
	browser, err := pw.Chromium.LaunchPersistentContext("/tmp/chatgpt", playwright.BrowserTypeLaunchPersistentContextOptions{Headless: playwright.Bool(headless)})
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Couldn't launch headless browser: %v", err))
	}
	page, err := browser.NewPage()
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Couldn't create a new tab on headless browser: %v", err))
	}

	if _, err = page.Goto(url); err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Couldn't open website: %v", err))
	}

	return browser, page, nil
}

type Result struct {
	Error        error
	SessionToken string
}

func logIn(pw *playwright.Playwright) <-chan Result {
	var lock sync.Mutex
	r := make(chan Result)

	lock.Lock()
	go func() {
		defer close(r)
		defer lock.Unlock()

		browser, page, err := launchBrowser(pw, "https://chat.openai.com/", false)
		if err != nil {
			r <- Result{Error: errors.New(fmt.Sprintf("Couldn't launch headless browser: %v", err))}
			return
		}
		log.Println("Please log in to OpenAI Chat")

		page.On("framenavigated", func(frame playwright.Frame) {
			if frame.URL() != "https://chat.openai.com/chat" {
				return
			}

			lock.Unlock()
		})

		lock.Lock()

		sessionToken, err := getSessionCookie(browser)
		if err != nil {
			r <- Result{Error: errors.New(fmt.Sprintf("Couldn't get session token: %v", err))}
			return
		}

		if err := browser.Close(); err != nil {
			r <- Result{Error: errors.New(fmt.Sprintf("Couldn't close headless browser: %v", err))}
			return
		}

		r <- Result{SessionToken: sessionToken}
	}()

	return r
}

func getSessionCookie(browser playwright.BrowserContext) (string, error) {
	cookies, err := browser.Cookies("https://chat.openai.com/")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Couldn't get cookies: %v", err))
	}

	var sessionToken string
	for _, cookie := range cookies {
		if cookie.Name == "__Secure-next-auth.session-token" {
			sessionToken = cookie.Value
			break
		}
	}

	if sessionToken == "" {
		return "", errors.New(fmt.Sprintf("Couldn't get session token"))
	}

	return sessionToken, nil
}
