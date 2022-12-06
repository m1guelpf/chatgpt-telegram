package chatgpt

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/m1guelpf/chatgpt-telegram/src/config"
	"github.com/m1guelpf/chatgpt-telegram/src/expirymap"
	"github.com/m1guelpf/chatgpt-telegram/src/sse"
)

const KEY_ACCESS_TOKEN = "accessToken"
const USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"

type ChatGPT struct {
	SessionToken   string
	AccessTokenMap expirymap.ExpiryMap
}

type SessionResult struct {
	Error       string `json:"error"`
	Expires     string `json:"expires"`
	AccessToken string `json:"accessToken"`
}

type MessageResponse struct {
	ConversationId string `json:"conversation_id"`
	Error          string `json:"error"`
	Message        struct {
		ID      string `json:"id"`
		Content struct {
			Parts []string `json:"parts"`
		} `json:"content"`
	} `json:"message"`
}

type ChatResponse struct {
	Message        string
	MessageId      string
	ConversationId string
}

func Init(config config.Config) ChatGPT {
	return ChatGPT{
		AccessTokenMap: expirymap.New(),
		SessionToken:   config.OpenAISession,
	}
}

func (c *ChatGPT) IsAuthenticated() bool {
	_, err := c.refreshAccessToken()
	return err == nil
}

func (c *ChatGPT) EnsureAuth() error {
	_, err := c.refreshAccessToken()
	return err
}

func (c *ChatGPT) SendMessage(message string, conversationId string, messageId string) (chan ChatResponse, error) {
	r := make(chan ChatResponse)
	accessToken, err := c.refreshAccessToken()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't get access token: %v", err))
	}

	client := sse.Init("https://chat.openai.com/backend-api/conversation")

	client.Headers = map[string]string{
		"User-Agent":    USER_AGENT,
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	err = client.Connect(message, conversationId, messageId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't connect to ChatGPT: %v", err))
	}

	go func() {
		defer close(r)
	mainLoop:
		for {
			select {
			case chunk, ok := <-client.EventChannel:
				if !ok {
					break mainLoop
				}

				var res MessageResponse
				err := json.Unmarshal([]byte(chunk), &res)
				if err != nil {
					log.Printf("Couldn't unmarshal message response: %v", err)
					continue
				}

				if len(res.Message.Content.Parts) > 0 {
					r <- ChatResponse{
						MessageId:      res.Message.ID,
						ConversationId: res.ConversationId,
						Message:        res.Message.Content.Parts[0],
					}
				}
			}
		}
	}()

	return r, nil
}

func (c *ChatGPT) refreshAccessToken() (string, error) {
	cachedAccessToken, ok := c.AccessTokenMap.Get(KEY_ACCESS_TOKEN)
	if ok {
		return cachedAccessToken, nil
	}

	req, err := http.NewRequest("GET", "https://chat.openai.com/api/auth/session", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Set("Cookie", fmt.Sprintf("__Secure-next-auth.session-token=%s", c.SessionToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %v", err)
	}
	defer res.Body.Close()

	var result SessionResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	accessToken := result.AccessToken
	if accessToken == "" {
		return "", errors.New("unauthorized")
	}

	if result.Error != "" {
		if result.Error == "RefreshAccessTokenError" {
			return "", errors.New("Session token has expired")
		}

		return "", errors.New(result.Error)
	}

	expiryTime, err := time.Parse(time.RFC3339, result.Expires)
	if err != nil {
		return "", fmt.Errorf("failed to parse expiry time: %v", err)
	}
	c.AccessTokenMap.Set(KEY_ACCESS_TOKEN, accessToken, expiryTime.Sub(time.Now()))

	return accessToken, nil
}
