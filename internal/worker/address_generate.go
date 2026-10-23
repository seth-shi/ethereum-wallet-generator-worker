package worker

import (
	"os"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func deriveAddress(seed []byte) (common.Address, error) {
	key, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return common.Address{}, err
	}

	// Preserve hdwallet v0.1.2's derivation behavior, including its optional
	// leading-zero fix. Changing recovery semantics requires a separate migration.
	fixIssue172 := len(os.Getenv("GO_ETHEREUM_HDWALLET_FIX_ISSUE_179")) > 0
	for _, child := range accounts.DefaultBaseDerivationPath {
		if fixIssue172 && key.IsAffectedByIssue172() {
			key, err = key.Derive(child)
		} else {
			key, err = key.DeriveNonStandard(child)
		}
		if err != nil {
			return common.Address{}, err
		}
	}

	publicKey, err := key.ECPubKey()
	if err != nil {
		return common.Address{}, err
	}
	hash := crypto.Keccak256(publicKey.SerializeUncompressed()[1:])
	return common.BytesToAddress(hash[12:]), nil
}
