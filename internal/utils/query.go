package utils

import (
	"errors"
	"net/url"
	"strings"

	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
)

func ParseQueryKey(fullUrl string) (string, string, error) {
	urlObj, err := url.Parse(fullUrl)
	if err != nil {
		return "", "", err
	}

	key := strings.TrimSpace(urlObj.Query().Get(consts.QueryKeyFieldName))
	if key == "" {
		return "", "", errors.New("服务端URL未包含秘钥")
	}

	if len(key) != consts.KeyLength {
		return "", "", errors.New("无效的秘钥,必须是16位")
	}

	urlObj.RawQuery = ""
	return urlObj.String(), key, nil
}
