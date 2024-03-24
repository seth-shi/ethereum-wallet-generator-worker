package master

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
)

type RunConfig struct {
	Version string
	// 数据保存
	FilePoint *os.File
	// 线程数量
	key string

	// 运行配置
	Port    int
	StartAt time.Time
}

func newRunConfig(port int, key string, startAt time.Time) (*RunConfig, error) {
	// 打开或创建一个csv文件，以追加模式写入
	pf, err := os.OpenFile(
		consts.MasterWalletDataFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0666,
	)
	if err != nil {
		return nil, err
	}

	key = strings.TrimSpace(key)
	if key == "" {
		key = lo.RandomString(16, lo.LowerCaseLettersCharset)
	}

	if len(key) != consts.KeyLength {
		return nil, errors.New("无效的秘钥,必须是16位")
	}

	return &RunConfig{
		Version:   utils.GetBuildVersion(),
		FilePoint: pf,
		key:       key,
		Port:      port,
		StartAt:   startAt,
	}, nil
}

func (rc *RunConfig) storeWalletData(address string, data string) error {

	// 创建一个csv写入器
	writer := csv.NewWriter(rc.FilePoint)
	// 循环写入数据
	err := writer.Write([]string{address, data})
	if err != nil {
		return errors.New(fmt.Sprintf("写入失败:[%s,%s]%s", address, data, err.Error()))
	}
	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
	return nil
}
