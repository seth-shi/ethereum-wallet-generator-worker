package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	tm "github.com/buger/goterm"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
)

type Worker struct {
	// 配置
	matchConfig  *models.MatchConfig
	runConfig    *RunConfig
	runStatus    *RunStatus
	outputString atomic.Pointer[string]
	httpClient   *resty.Client
}

func NewWorker(fullUrl string, mc *models.MatchConfig, c uint, nodeName string) (*Worker, error) {

	runConfig, err := newRunConfig(fullUrl, c, nodeName)
	if err != nil {
		return nil, err
	}

	return &Worker{
		matchConfig:  mc,
		runConfig:    runConfig,
		runStatus:    newRunStatus(),
		outputString: atomic.Pointer[string]{},
		httpClient:   resty.New().SetTimeout(time.Second * 3),
	}, nil
}

func (w *Worker) Run() error {

	// 启动上报一次
	for i := 0; i < w.runConfig.C; i++ {
		go w.loopMatchWallets()
	}

	// 定时上报状态
	go w.timerReportServer()

	// 刷新输出
	w.timerOutput()

	return w.runConfig.FilePoint.Close()
}

func (w *Worker) timerReportServer() {

	// 上报时长 5s
	utils.MustError(w.reportServer(nil))
	timer := time.NewTicker(time.Second * 5)
	for range timer.C {

		if err := w.reportServer(nil); err != nil {
			fmt.Println(err)
		}
	}
}

func (w *Worker) loopMatchWallets() {

	for {

		newWalletData := w.runStatus.matchNewWallet(w.matchConfig)
		if newWalletData != nil {
			if err := w.reportServer(newWalletData); err != nil {
				if storeErr := w.runConfig.storeWalletData(newWalletData); storeErr != nil {
					utils.MustError(storeErr)
				}
			}
		}
	}
}

func (w *Worker) timerOutput() {
	timer := time.NewTicker(time.Second)
	tm.Clear()
	var lastMinute = time.Now().Minute()
	for ts := range timer.C {

		nowMinute := ts.Minute()
		if nowMinute > lastMinute {
			lastMinute = nowMinute
			tm.Clear()
		}

		// 永远返回不失败
		tm.MoveCursor(0, 0)
		_, _ = tm.Println(strings.Repeat("-", consts.LineCharCount))
		tm.MoveCursor(0, 2)
		_, _ = tm.Println(fmt.Sprintf("--版本号:%s", w.runConfig.Version))
		_, _ = tm.Println(fmt.Sprintf("--节点名:%s 线程*%d", w.runConfig.Name, w.runConfig.C))
		_, _ = tm.Println(fmt.Sprintf(
			"--实时速度: %.2f 钱包/秒 生成:%d 找到:%d",
			w.runStatus.Speed(),
			w.runStatus.TotalCount.Load(),
			w.runStatus.FoundCount.Load(),
		))
		_, _ = tm.Println(*w.outputString.Load())
		tm.Flush()
	}
}

func (w *Worker) reportServer(wa *models.Wallet) (err error) {

	recentCount := w.runStatus.RecentCount.Swap(0)
	defer func() {
		if err != nil {
			// 恢复数量, 中途可能数量增加了
			w.runStatus.RecentCount.Add(recentCount)
		}
	}()

	// 计算时间
	progressReq := &models.WorkStatusRequest{
		Name:         w.runConfig.Name,
		BuildVersion: w.runConfig.Version,
		Count:        int(recentCount),
		Found:        int(w.runStatus.FoundCount.Load()),
		Speed:        w.runStatus.Speed(),
		StartAt:      w.runStatus.StartAt,
	}
	if wa != nil {
		encryptData, err := utils.AesGcmEncrypt(wa.Mnemonic, w.runConfig.key)
		if err != nil {
			return err
		}
		progressReq.Address = lo.ToPtr(wa.Address)
		progressReq.EncryptMnemonic = lo.ToPtr(encryptData)
	}

	data, err := json.Marshal(progressReq)
	if err != nil {
		return err
	}

	resp, err := w.httpClient.R().SetBody(data).Post(w.runConfig.MasterHost)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New(fmt.Sprintf("http post status[%s][%s]", resp.Status(), resp.String()))
	}

	bodyContent, err := url.QueryUnescape(resp.String())
	if err != nil {
		return err
	}

	bodyContent = strings.Trim(bodyContent, "\"")
	w.outputString.Swap(&bodyContent)

	return nil
}
