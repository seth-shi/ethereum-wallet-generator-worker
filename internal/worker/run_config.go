package worker

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"os"
	"strings"

	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
)

type RunConfig struct {
	Name    string
	Version string
	// 服务端地址
	MasterHost string
	// 线程数量
	C int

	key []byte
}

func newRunConfig(fullUrl string, c uint, name string) (*RunConfig, error) {

	return &RunConfig{
		Name:       name,
		Version:    utils.GetBuildVersion(),
		MasterHost: fullUrl,
		C:          int(c),
		key:        []byte(lo.RandomString(consts.KeyLength, lo.LowerCaseLettersCharset)),
	}, nil
}

func (rc *RunConfig) storeWalletData(wa *models.Wallet) {

	// 凡是出错, 直接打印原始出来在标准输出
	// node 保存钱包的时候, 也需要加密数据
	encryptData, err := utils.AesGcmEncrypt(wa.Mnemonic, rc.key)
	if err != nil {
		utils.MustError(errors.New(fmt.Sprintf("钱包加密失败:[%s,%s]%s", wa.Address, wa.Mnemonic, err.Error())))
	}

	line := []string{wa.Address, string(rc.key), encryptData}
	lineStr := strings.Join(line, ",")
	// 打开或创建一个csv文件，以追加模式写入
	pf, err := os.OpenFile(
		consts.WorkerWalletDataFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0666,
	)
	if err != nil {
		utils.MustError(errors.New(fmt.Sprintf("打开文件失败:[%s]%s", lineStr, err.Error())))
	}

	// 创建一个csv写入器
	writer := csv.NewWriter(pf)
	// 循环写入数据
	if err := writer.Write(line); err != nil {
		utils.MustError(errors.New(fmt.Sprintf("钱包写入失败:[%s]%s", lineStr, err.Error())))
	}

	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
}
