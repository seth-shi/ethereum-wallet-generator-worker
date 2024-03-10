package internal

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
)

type Wallet struct {
	Address  string `json:"address"`
	Mnemonic []byte `json:"mnemonic"`
}

func newWallet() (*Wallet, error) {
	// 生成随机熵（128位）
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}

	// 根据熵生成助记词
	mnemonic, _ := bip39.NewMnemonic(entropy)

	// 使用助记词生成种子
	seed := bip39.NewSeed(mnemonic, "")

	// 创建 HD 钱包
	hdWallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		return nil, err
	}

	// 指定路径生成 ETH 地址
	account, err := hdWallet.Derive(accounts.DefaultBaseDerivationPath, false)
	if err != nil {
		return nil, err
	}

	// 获取地址、私钥和公钥
	address := account.Address.Hex()
	return &Wallet{Address: address, Mnemonic: []byte(mnemonic)}, nil
}
