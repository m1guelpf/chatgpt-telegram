package chatgpt

import (
	"context"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type IChatClient interface {
	Talk(ctx context.Context, req *gogpt.CompletionRequest) (result *gogpt.CompletionResponse, err error)
}

const (
	DefaultModel       = "text-davinci-003"
	DefaultMaxTokens   = 150
	DefaultTemperature = 0.5
	DefaultTop         = 1
	DefaultN           = 1
	DefaultBest        = 1
	DefaultUser        = "default"
	DefaultStream      = false

	DefaultZero = 0

	NoValueString = ""
	NoValueInt    = 0
	NoValueFloat  = 0.0
)

type chatGptClientImpl struct {
	client *gogpt.Client
}

func NewChatGptClient(token string) IChatClient {
	return &chatGptClientImpl{client: gogpt.NewClient(token)}
}

func (impl *chatGptClientImpl) Talk(ctx context.Context, req *gogpt.CompletionRequest) (*gogpt.CompletionResponse, error) {
	impl.checkReq(req)
	result, err := impl.client.CreateCompletion(ctx, *req)
	return &result, err
}

func (impl *chatGptClientImpl) checkReq(req *gogpt.CompletionRequest) {
	if req.Model == NoValueString {
		req.Model = DefaultModel
	}
	if req.MaxTokens == NoValueInt {
		req.MaxTokens = DefaultMaxTokens
	}
	if req.Temperature == NoValueFloat {
		req.Temperature = DefaultTemperature
	}
	if req.User == NoValueString {
		req.User = DefaultUser
	}
	if req.TopP == NoValueFloat {
		req.TopP = DefaultTop
	}
	if req.N == NoValueInt {
		req.N = DefaultN
	}
	if req.BestOf == NoValueInt {
		req.BestOf = DefaultBest
	}
	req.Stream = DefaultStream
	req.Echo = false
	req.PresencePenalty = DefaultZero
	req.FrequencyPenalty = DefaultZero
}

func NewDefaultCompletionRequest(prompt string, user string) *gogpt.CompletionRequest {
	return &gogpt.CompletionRequest{
		Prompt:    prompt,
		Suffix:    "",
		LogProbs:  0,
		Stop:      nil,
		LogitBias: nil,
		User:      user,
	}
}
