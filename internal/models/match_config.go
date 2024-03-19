package models

import (
	"fmt"
	"math"
	"strings"

	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
)

type MatchConfig struct {
	Prefix   string `json:"prefix"`
	Suffix   string `json:"suffix"`
	MayCount int64  `json:"-"`
}

func NewMatchConfig(prefix, suffix string) *MatchConfig {

	matchLength := len(suffix)
	if prefix != "" {
		matchLength += len(prefix)
		if !strings.HasPrefix(prefix, consts.AddressPrefix) {
			prefix = fmt.Sprintf("%s%s", consts.AddressPrefix, prefix)
		} else {
			matchLength -= len(consts.AddressPrefix)
		}
	}

	return &MatchConfig{
		Prefix:   prefix,
		Suffix:   suffix,
		MayCount: int64(math.Pow(16, float64(matchLength))),
	}
}
