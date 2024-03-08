package internal

import "time"

const (
	addressPrefix = "0x"
	keyFieldName  = "key"
	lineCharCount = 80
	keyLength     = 16
)

type GetConfigRequest struct {
	MayCount uint64 `json:"-"`
	Prefix   string `json:"prefix"`
	Suffix   string `json:"suffix"`
}

type NodeProgress struct {
	Name       string  `json:"name"`
	Count      int     `json:"gen_count"`
	Found      int     `json:"found_count"`
	Speed      float64 `json:"speed"`
	WalletData *Wallet `json:"wallet_data"`

	LastActiveAt time.Time `json:"-"`
}
