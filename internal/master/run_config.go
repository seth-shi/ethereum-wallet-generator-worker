package master

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"os"
	"strings"
	"time"

	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
)

type RunConfig struct {
	Version string
	// 数据保存
	FilePoint *os.File
	// 线程数量

	// 运行配置
	Port    int
	StartAt time.Time
}

func newRunConfig(port int, startAt time.Time) (*RunConfig, error) {
	// 打开或创建一个csv文件，以追加模式写入
	// 启动时候打开, 放于后续生成成功的时候打不开
	pf, err := os.OpenFile(
		consts.MasterWalletDataFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0666,
	)
	if err != nil {
		return nil, err
	}

	return &RunConfig{
		Version:   utils.GetBuildVersion(),
		FilePoint: pf,
		Port:      port,
		StartAt:   startAt,
	}, nil
}

func (rc *RunConfig) storeWalletData(wa *models.WalletModel) error {

	// 创建一个csv写入器
	line := []string{wa.Address, wa.Key, wa.EncryptMnemonic}
	writer := csv.NewWriter(rc.FilePoint)
	// 循环写入数据
	err := writer.Write(line)
	if err != nil {
		return errors.New(fmt.Sprintf("写入失败:[%s]%s", strings.Join(line, ","), err.Error()))
	}
	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
	return nil
}
