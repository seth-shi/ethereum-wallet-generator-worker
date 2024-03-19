package worker

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal/models"
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal/utils"
	"os"
)

type RunConfig struct {
	Name    string
	Version string
	// 数据保存
	FilePoint *os.File
	// 服务端地址
	MasterHost string
	// 线程数量
	C int

	key []byte
}

func newRunConfig(fullUrl string, c uint, name string) (*RunConfig, error) {
	// 打开或创建一个csv文件，以追加模式写入
	pf, err := os.OpenFile(
		consts.WorkerWalletDataFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0666,
	)
	if err != nil {
		return nil, err
	}

	host, key, err := utils.ParseQueryKey(fullUrl)
	if err != nil {
		return nil, err
	}

	return &RunConfig{
		Name:       name,
		Version:    utils.GetBuildVersion(),
		FilePoint:  pf,
		MasterHost: host,
		C:          int(c),
		key:        []byte(key),
	}, nil
}

func (rc *RunConfig) storeWalletData(wa *models.Wallet) error {

	// 凡是出错, 直接打印原始出来在标准输出
	// node 保存钱包的时候, 也需要加密数据
	encryptData, err := utils.AesGcmEncrypt(wa.Mnemonic, rc.key)
	if err != nil {
		return errors.New(fmt.Sprintf("钱包加密失败:[%s,%s]%s", wa.Address, wa.Mnemonic, err.Error()))
	}

	// 创建一个csv写入器
	writer := csv.NewWriter(rc.FilePoint)
	// 循环写入数据
	if err := writer.Write([]string{wa.Address, encryptData}); err != nil {
		return errors.New(fmt.Sprintf("钱包写入失败:[%s,%s]%s", wa.Address, wa.Mnemonic, err.Error()))
	}

	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
	return nil
}
