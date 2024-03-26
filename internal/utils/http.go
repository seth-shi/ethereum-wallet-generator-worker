package utils

import (
	"github.com/go-resty/resty/v2"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"time"
)

var (
	defaultClient = resty.New().SetTimeout(time.Second * 3)
)

func IPV4() string {

	resp, err := defaultClient.R().Get("https://api.ipify.org")
	if err != nil {
		return ""
	}

	return resp.String()
}

func GetMatchConfig(server string) (*models.MatchConfig, *resty.Response, error) {
	var vc *models.MatchConfig
	resp, err := defaultClient.R().SetResult(vc).Get(server)
	return vc, resp, err
}
