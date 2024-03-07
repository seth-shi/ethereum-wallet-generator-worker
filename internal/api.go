package internal

type GetConfigRequest struct {
	MayCount uint64 `json:"-"`
	Prefix   string `json:"prefix"`
	Suffix   string `json:"suffix"`
	Key      string `json:"key"`
}

type NodeProgress struct {
	Name       string  `json:"name"`
	Count      int     `json:"gen_count"`
	Found      int     `json:"found_count"`
	Speed      float64 `json:"speed"`
	WalletData *Wallet `json:"wallet_data"`
}
