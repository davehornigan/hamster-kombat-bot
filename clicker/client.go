package clicker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/davehornigan/hamster-kombat-bot/clicker/boost"
	"github.com/davehornigan/hamster-kombat-bot/clicker/tap"
	"github.com/davehornigan/hamster-kombat-bot/clicker/upgrades"
	"github.com/davehornigan/hamster-kombat-bot/clicker/user"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	DefaultBaseURL = "https://api.hamsterkombatgame.io/clicker"
	SyncUri        = "/sync"
	TapUri         = "/tap"
	BoostsUri      = "/upgrades-for-buy"
)

type RequestInterface interface {
	IsRequest() bool
}

type ResponseInterface interface {
	IsResponse() bool
}

type Response struct {
	ClickerUser *user.User `json:"clickerUser"`
}

func (r *Response) IsResponse() bool {
	return true
}

type Client struct {
	baseURL    string `env:"CLIENT_BASE_URL" envDefault:"https://api.hamsterkombatgame.io/clicker"`
	httpClient *http.Client
	headers    *http.Header
}

func NewClient(authToken string, userAgent string) *Client {
	baseUrl := DefaultBaseURL
	if os.Getenv("CLIENT_BASE_URL") != "" {
		baseUrl = os.Getenv("CLIENT_BASE_URL")
	}
	client := &Client{
		baseURL:    baseUrl,
		httpClient: http.DefaultClient,
		headers: &http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", authToken)},
			"User-Agent":    []string{userAgent},
			"Content-Type":  []string{"application/json"},
			"Accept":        []string{"application/json"},
		},
	}

	return client
}

func (c *Client) Sync() (*user.User, error) {
	res, err := sendRequest[*Response](c, SyncUri, nil, &Response{})
	if err != nil {
		return nil, err
	}
	return res.ClickerUser, nil
}

func (c *Client) Tap(count int32, available int32) (*user.User, error) {
	res, err := sendRequest[*Response](c, TapUri, &tap.Request{
		Count:         count,
		Timestamp:     time.Now().Unix(),
		AvailableTaps: available,
	}, &Response{})
	if err != nil {
		return nil, err
	}
	return res.ClickerUser, nil
}

func (c *Client) CheckUpgrades() (*upgrades.Response, error) {
	return sendRequest[*upgrades.Response](c, upgrades.ForBuyUri, nil, &upgrades.Response{})
}

func (c *Client) BuyUpgrade(upgrade *upgrades.Upgrade) (*upgrades.Response, error) {
	return sendRequest[*upgrades.Response](c, upgrades.BuyUri, &upgrades.BuyUpgrade{
		UpgradeId: upgrade.ID,
		Timestamp: time.Now().Unix(),
	}, &upgrades.Response{})
}

func (c *Client) GetBoostsForBuy() ([]*boost.Boost, error) {
	boostsForBuy, err := sendRequest(c, boost.ForBuyUri, nil, &boost.BoostsForBuy{})
	if err != nil {
		return nil, err
	}

	return boostsForBuy.Boosts, nil
}

func (c *Client) BuyBoost(boostForBuy *boost.Boost) ([]*boost.Boost, error) {
	boostsForBuy, err := sendRequest(c, boost.BuyUri, &boost.Buy{
		BoostId:   boostForBuy.Id,
		Timestamp: time.Now().Unix(),
	}, &boost.BoostsForBuy{})
	if err != nil {
		return nil, err
	}

	return boostsForBuy.Boosts, nil
}

func sendRequest[RP ResponseInterface](c *Client, uri string, reqStruct RequestInterface, resStruct RP) (RP, error) {
	var reqBytes []byte
	var err error
	if reqStruct != nil {
		reqBytes, err = json.Marshal(reqStruct)
		if err != nil {
			panic(err)
		}
	}
	reqUrl := c.baseURL + uri
	reqCloser := io.NopCloser(bytes.NewReader([]byte(``)))
	if reqBytes != nil {
		reqCloser = io.NopCloser(bytes.NewReader(reqBytes))
	}
	req, err := http.NewRequest("POST", reqUrl, reqCloser)
	if err != nil {
		return resStruct, fmt.Errorf("client: could not create request: %s\n", err)
	}
	req.Header = *c.headers

	res, err := c.httpClient.Do(req)
	if err != nil {
		return resStruct, fmt.Errorf("client: could not make request: %s\n", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return resStruct, fmt.Errorf("client: could not read response body: %s\n", err)
	}

	if err := json.Unmarshal(body, resStruct); err != nil {
		return resStruct, fmt.Errorf("client: could not parse response body: %s\n", err)
	}
	if res.StatusCode >= 400 {
		return resStruct, fmt.Errorf("client: could not send request: %s\ncode: %d\n", body, res.StatusCode)
	}

	return resStruct, nil
}
