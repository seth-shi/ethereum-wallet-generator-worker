package models

type WalletModel struct {
	Address         string `json:"address"`
	EncryptMnemonic string `json:"mnemonic"`
	Key             string `json:"key"`
}
