package worker

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"github.com/tyler-smith/go-bip39"
)

func BenchmarkMatchNewWallet(b *testing.B) {
	config := &models.MatchConfig{Prefix: "0x0000000000000000000000000000000000000000"}
	for _, implementation := range []struct {
		name     string
		generate func(*RunStatus, *models.MatchConfig) (*models.WalletModel, error)
	}{
		{"reference", (*RunStatus).matchNewWalletReference},
		{"optimized", (*RunStatus).matchNewWallet},
	} {
		b.Run(implementation.name+"/serial", func(b *testing.B) {
			status := newRunStatus()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := implementation.generate(status, config); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(implementation.name+"/parallel", func(b *testing.B) {
			status := newRunStatus()
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if _, err := implementation.generate(status, config); err != nil {
						b.Error(err)
						return
					}
				}
			})
		})
	}
}

func BenchmarkGenerationStages(b *testing.B) {
	mnemonic, err := bip39.NewMnemonic(make([]byte, 16))
	if err != nil {
		b.Fatal(err)
	}
	seed := bip39.NewSeed(mnemonic, "")
	b.Run("seed", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = bip39.NewSeed(mnemonic, "")
		}
	})
	b.Run("derive", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			wallet, err := hdwallet.NewFromSeed(seed)
			if err != nil {
				b.Fatal(err)
			}
			if _, err := wallet.Derive(accounts.DefaultBaseDerivationPath, false); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("deriveOptimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := deriveAddress(seed); err != nil {
				b.Fatal(err)
			}
		}
	})
}
