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

	"github.com/go-resty/resty/v2"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
)

type Worker struct {
	// 配置
	Title        string
	matchConfig  *models.MatchConfig
	runConfig    *RunConfig
	runStatus    *RunStatus
	outputString atomic.Pointer[string]
	httpClient   *resty.Client
}

func NewWorker(fullUrl string, mc *models.MatchConfig, c uint, workerName string) (*Worker, error) {

	runConfig, err := newRunConfig(fullUrl, c, workerName)
	if err != nil {
		return nil, err
	}

	return &Worker{
		Title:        fmt.Sprintf("--版本号:%s\n--节点名:%s 线程*%d", runConfig.Version, runConfig.Name, runConfig.C),
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

	return nil
}

func (w *Worker) timerReportServer() {

	// 上报时长 5s
	utils.MustError(w.reportServer(nil))
	for range time.Tick(time.Second * 2) {
		utils.ShowIfError(w.reportServer(nil))
	}
}

func (w *Worker) loopMatchWallets() {

	for {
		newWalletData, err := w.runStatus.matchNewWallet(w.matchConfig)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if newWalletData != nil {
			if err := w.reportServer(newWalletData); err != nil {
				w.runConfig.storeWalletData(newWalletData)
			}
		}
	}
}

func (w *Worker) timerOutput() {
	var lastMinute = time.Now().Minute()
	for ts := range time.Tick(time.Second) {

		nowMinute := ts.Minute()
		if nowMinute > lastMinute {
			lastMinute = nowMinute
			fmt.Print("\033[2J\033[H")
		}

		fmt.Printf(
			"\u001B[H%s\n%s\n%s\n%s\n%s",
			strings.Repeat("-", consts.LineCharCount),
			w.Title,
			strings.Repeat("-", consts.LineCharCount),
			fmt.Sprintf(
				"--实时速度: %.2f 钱包/秒 生成:%d 找到:%d",
				w.runStatus.Speed(),
				w.runStatus.TotalCount.Load(),
				w.runStatus.FoundCount.Load(),
			),
			*w.outputString.Load(),
		)
	}
}

func (w *Worker) reportServer(wa *models.WalletModel) (err error) {

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
		Speed:        w.runStatus.Speed(),
		StartAt:      w.runStatus.StartAt,
		Wallet:       wa,
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
