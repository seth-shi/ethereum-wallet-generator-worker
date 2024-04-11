package master

import "github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"

var (
	CsvHeaders = &models.WalletModel{
		Address:         "公钥",
		EncryptMnemonic: "加密后的助记词(请查看readme解密)",
		Key:             "加密Key",
	}
)
