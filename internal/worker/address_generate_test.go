package worker

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func referenceAddress(seed []byte) (common.Address, error) {
	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		return common.Address{}, err
	}
	account, err := wallet.Derive(accounts.DefaultBaseDerivationPath, false)
	return account.Address, err
}

func TestDeriveAddressCompatibility(t *testing.T) {
	indices := make([]int, 128, 130)
	for i := range indices {
		indices[i] = i
	}
	// These also cover short private scalars before hardened/non-hardened
	// children and at the final key, including the historical BIP-32 issue.
	indices = append(indices, 182, 1028)
	for _, setting := range []string{"", "1", "0"} {
		t.Run("fix="+setting, func(t *testing.T) {
			t.Setenv("GO_ETHEREUM_HDWALLET_FIX_ISSUE_179", setting)
			for _, i := range indices {
				first := sha256.Sum256([]byte(fmt.Sprintf("wallet-derive-compat/%d/0", i)))
				second := sha256.Sum256([]byte(fmt.Sprintf("wallet-derive-compat/%d/1", i)))
				seed := append(first[:], second[:]...)
				want, err := referenceAddress(seed)
				if err != nil {
					t.Fatal(err)
				}
				got, err := deriveAddress(seed)
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Fatalf("fixture %d: address differs from original implementation", i)
				}
			}
		})
	}
}
