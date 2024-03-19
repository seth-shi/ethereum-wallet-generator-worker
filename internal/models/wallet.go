package models

type Wallet struct {
	Address  string `json:"address"`
	Mnemonic []byte `json:"mnemonic"`
}
