package worker

import (
	"github.com/samber/lo"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"github.com/tyler-smith/go-bip39"
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

func (r *RunStatus) Speed() float64 {
	var speed = 0.0
	diff := time.Now().Unix() - r.StartAt
	if diff > 0 {
		speed = float64(r.TotalCount.Load()) / float64(diff)
	}
	return speed
}

func (r *RunStatus) matchNewWallet(matchConfig *models.MatchConfig) (*models.WalletModel, error) {

	defer func() {
		r.TotalCount.Add(1)
		r.RecentCount.Add(1)
	}()

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

	if matchConfig.Prefix != "" && !strings.HasPrefix(address, matchConfig.Prefix) {
		return nil, nil
	}

	if matchConfig.Suffix != "" && !strings.HasSuffix(address, matchConfig.Suffix) {
		return nil, nil
	}

	// 加密
	key := lo.RandomString(consts.KeyLength, lo.LowerCaseLettersCharset)
	encryptData, err := utils.AesGcmEncrypt([]byte(mnemonic), []byte(key))
	if err != nil {
		return nil, err
	}

	r.FoundCount.Add(1)
	return &models.WalletModel{Address: address, EncryptMnemonic: encryptData, Key: key}, nil
}
