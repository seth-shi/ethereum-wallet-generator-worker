package worker

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal/models"
	"github.com/tyler-smith/go-bip39"
	"strings"
	"sync/atomic"
	"time"
)

type RunStatus struct {
	TotalCount  atomic.Uint64
	FoundCount  atomic.Int32
	RecentCount atomic.Int32

	StartAt int64
}

func newRunStatus() *RunStatus {
	return &RunStatus{
		TotalCount:  atomic.Uint64{},
		FoundCount:  atomic.Int32{},
		RecentCount: atomic.Int32{},
		StartAt:     time.Now().Unix(),
	}
}

func (r *RunStatus) matchNewWallet(matchConfig *models.MatchConfig) *models.Wallet {
	defer func() {
		r.TotalCount.Add(1)
		r.RecentCount.Add(1)
	}()

	wallet, err := r.newWallet()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if matchConfig.Prefix != "" && !strings.HasPrefix(wallet.Address, matchConfig.Prefix) {
		return nil
	}

	if matchConfig.Suffix != "" && !strings.HasSuffix(wallet.Address, matchConfig.Suffix) {
		return nil
	}

	r.FoundCount.Add(1)

	return wallet
}

func (r *RunStatus) Speed() float64 {
	var speed = 0.0
	diff := time.Now().Unix() - r.StartAt
	if diff > 0 {
		speed = float64(r.TotalCount.Load()) / float64(diff)
	}
	return speed
}

func (r *RunStatus) newWallet() (*models.Wallet, error) {
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
	return &models.Wallet{Address: address, Mnemonic: []byte(mnemonic)}, nil
}
