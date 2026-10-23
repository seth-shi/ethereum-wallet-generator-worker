package worker

import (
	"github.com/samber/lo"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
	"strings"
	"sync/atomic"
	"time"

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
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	// 使用助记词生成种子
	seed := bip39.NewSeed(mnemonic, "")

	// Derive only the address; no wallet/account wrapper is needed per candidate.
	accountAddress, err := deriveAddress(seed)
	if err != nil {
		return nil, err
	}

	if !mayMatchAddress(accountAddress, matchConfig.Prefix, matchConfig.Suffix) {
		return nil, nil
	}
	// Only possible matches need the EIP-55 checksum and exact case check.
	address := accountAddress.Hex()

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
