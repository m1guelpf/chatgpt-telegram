package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/shamu00/chatgpt-telegram/src/util"
)

const (
	KeyTelegramId     = "CHAT_GPT_BOT_TELEGRAM_ID"
	KeyTelegramToken  = "CHAT_GPT_BOT_TELEGRAM_TOKEN"
	KeyChatGptOpenKey = "CHAT_GPT_OEPNAPI_KEY"

	// Insert Into AppService
	AzureConfigCenterCredential = "AZURE_CONFIG_CENTER_CREDENTIAL"
	AzureConfigCenterSecret     = "AZURE_CONFIG_CENTER_SECRET"

	PathGetFormat = "/kv/%s?api-version=1.0"

	invalidReturn = "error"
)

type GetResponse struct {
	Etag         *string           `json:"etag"`
	Key          *string           `json:"key"`
	Label        *string           `json:"label"`
	ContentType  *string           `json:"content_type"`
	Value        *string           `json:"value"`
	LastModified *time.Time        `json:"last_modified"`
	Locked       *bool             `json:"locked"`
	Tags         map[string]string `json:"tags"`
}

func (g *GetResponse) String() string {
	return fmt.Sprintf(`{etag:%s,key:%s,label:%v,content_type:%s,value:%s,last_modified:%v,locked:%v,tags:%v}`,
		ins(g.Etag), ins(g.Key), ins(g.Label), ins(g.ContentType), ins(g.Value), ins(g.LastModified), ins(g.Locked), g.Tags)
}

func ins(v any) any {
	if v == nil || reflect.ValueOf(v).IsNil() {
		return v
	}
	if reflect.TypeOf(v).Kind() == reflect.Pointer {
		return reflect.ValueOf(v).Elem().Interface()
	}
	return v
}

type IConfigurationFetcher interface {
	GetString(ctx context.Context, key string) (res string, err error)
}

type ConfigurationFetcherImpl struct {
	httpEndPoint    string
	httpClient      *http.Client
	azureCredential string
	azureSecret     string
}

var (
	keyPathMapper = map[string]string{}
)

func RegisterKeyPath(key, path string) {
	keyPathMapper[key] = path
}

func InitConfigurationFetcher() {
	RegisterKeyPath(KeyTelegramId, PathGetFormat)
	RegisterKeyPath(KeyTelegramToken, PathGetFormat)
	RegisterKeyPath(KeyChatGptOpenKey, PathGetFormat)
}

func NewAzureConfigurationFetcher(httpEndPoint, azureCredential, azureSecret string) IConfigurationFetcher {
	return &ConfigurationFetcherImpl{
		httpEndPoint:    httpEndPoint,
		httpClient:      &http.Client{},
		azureCredential: azureCredential,
		azureSecret:     azureSecret,
	}
}

func (impl *ConfigurationFetcherImpl) GetString(_ context.Context, key string) (string, error) {
	format, found := keyPathMapper[key]
	if !found {
		format = key
	}
	url := impl.httpEndPoint + fmt.Sprintf(format, key)
	var resp *http.Response
	var err error
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return invalidReturn, err
	}
	req.Body = io.NopCloser(strings.NewReader(""))

	err = util.SignRequest(impl.azureCredential, impl.azureSecret, req)
	if err != nil {
		return invalidReturn, err
	}
	err = util.Retry(3, 100*time.Millisecond, func() error {
		var e error
		resp, e = impl.httpClient.Do(req)
		return e
	})
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return invalidReturn, err
	}
	if resp.StatusCode != http.StatusOK {
		return invalidReturn, fmt.Errorf("http error, code:%d", resp.StatusCode)
	}
	bs, err := io.ReadAll(resp.Body)
	str := string(bs)
	_ = str
	if err != nil {
		return invalidReturn, err
	}
	var res = &GetResponse{}
	err = json.Unmarshal(bs, res)
	if err != nil {
		return invalidReturn, err
	}
	if res.Value == nil {
		return invalidReturn, nil
	}
	return *res.Value, nil
}
