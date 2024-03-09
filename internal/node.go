package internal

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	tm "github.com/buger/goterm"
	"github.com/go-resty/resty/v2"
	"github.com/golang-module/dongle"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Node struct {
	Name string

	// 总量
	TotalCount  atomic.Uint64
	FoundCount  atomic.Int32
	RecentCount atomic.Int32
	StartAt     int64

	// 数据保存
	FilePoint *os.File

	// 服务端地址
	Host string `json:"-"`

	// 线程数量
	C uint `json:"-"`

	// 加密解密
	Cip *dongle.Cipher

	// 配置
	Config GetConfigRequest

	// 请求
	HttpClient *resty.Client

	OutputString atomic.Pointer[string]
}

func NewNode(host string, cfg GetConfigRequest, c uint, nodeName string) (*Node, error) {

	// 打开或创建一个csv文件，以追加模式写入
	pf, err := os.OpenFile("wallet.node.csv", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	urlObj, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	key := strings.TrimSpace(urlObj.Query().Get(keyFieldName))
	if key == "" {
		return nil, errors.New("服务端URL未包含秘钥")
	}

	if len(key) != keyLength {
		return nil, errors.New("无效的秘钥,必须是16位")
	}

	urlObj.RawQuery = ""
	host = urlObj.String()

	return &Node{
		Host:       host,
		FilePoint:  pf,
		C:          c,
		Name:       nodeName,
		Cip:        getCipher(key),
		Config:     cfg,
		StartAt:    time.Now().Unix(),
		HttpClient: resty.New().SetTimeout(time.Second * 5),
	}, nil
}

func (n *Node) Run() {

	// 启动上报一次
	MustError(n.reportServer(nil))
	for i := 0; i < int(n.C); i++ {
		go n.loopMatchWallets()
	}

	// 定时上报状态
	go n.timerReportServer()

	// 刷新输出
	n.timerOutput()
}
func (n *Node) speed(nowUnix int64) float64 {
	var speed = 0.0
	diff := nowUnix - n.StartAt
	if diff > 0 {
		speed = float64(n.TotalCount.Load()) / float64(diff)
	}
	return speed
}
func (n *Node) timerOutput() {
	timer := time.NewTicker(time.Second)
	tm.Clear()
	tm.MoveCursor(0, 0)
	_, _ = tm.Println(strings.Repeat("-", lineCharCount))
	for ts := range timer.C {

		// 永远返回不失败
		tm.MoveCursor(0, 2)
		_, _ = tm.Println(fmt.Sprintf("--节点名:%s 线程*%d", n.Name, n.C))
		_, _ = tm.Println(fmt.Sprintf(
			"--实时速度: %.2f 钱包/秒 生成:%d 找到:%d",
			n.speed(ts.Unix()),
			n.TotalCount.Load(),
			n.FoundCount.Load(),
		))
		_, _ = tm.Println(*n.OutputString.Load())
		tm.Flush()
	}
}

func (n *Node) timerReportServer() {

	// 上报时长 5s
	timer := time.NewTicker(time.Second * 5)
	for range timer.C {

		if err := n.reportServer(nil); err != nil {
			fmt.Println(err)
		}
	}
}

func (n *Node) loopMatchWallets() {

	for {
		if newWalletData := n.matchNewWallet(); newWalletData != nil {
			if err := n.reportServer(newWalletData); err != nil {
				if err1 := n.storeWalletData(newWalletData); err1 != nil {
					fmt.Printf("\n%s,%s\n", newWalletData.Address, newWalletData.Mnemonic)
					MustError(err1)
				}
			}
		}
	}
}

func (n *Node) storeWalletData(data *Wallet) error {

	// 创建一个csv写入器
	writer := csv.NewWriter(n.FilePoint)
	// 循环写入数据
	err := writer.Write([]string{data.Address, data.Mnemonic})
	if err != nil {
		return errors.New(fmt.Sprintf("钱包写入失败:[%s,%s]%s", data.Address, data.Mnemonic, err.Error()))
	}
	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
	return nil
}

func (n *Node) matchNewWallet() *Wallet {
	defer func() {
		n.TotalCount.Add(1)
		n.RecentCount.Add(1)
	}()

	wallet, err := newWallet()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if n.Config.Prefix != "" && !strings.HasPrefix(wallet.Address, n.Config.Prefix) {
		return nil
	}

	if n.Config.Suffix != "" && !strings.HasSuffix(wallet.Address, n.Config.Suffix) {
		return nil
	}

	n.FoundCount.Add(1)

	return wallet
}

func (n *Node) reportServer(wa *Wallet) (err error) {

	nowUnix := time.Now().Unix()
	recentCount := n.RecentCount.Swap(0)
	defer func() {
		if err != nil {
			// 恢复数量, 中途可能数量增加了
			n.RecentCount.Add(recentCount)
		}
	}()

	// 计算时间
	progressReq := &NodeProgress{
		Name:       n.Name,
		Count:      int(recentCount),
		Found:      int(n.FoundCount.Load()),
		Speed:      n.speed(nowUnix),
		StartAt:    n.StartAt,
		WalletData: wa,
	}
	data, _ := json.Marshal(progressReq)
	encryptData := dongle.Encrypt.FromBytes(data).ByAes(n.Cip).ToRawBytes()
	resp, err := n.HttpClient.R().SetBody(encryptData).Post(n.Host)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New(fmt.Sprintf("http post status %s", resp.Status()))
	}

	bodyContent, err := url.QueryUnescape(resp.String())
	if err != nil {
		return err
	}

	bodyContent = strings.Trim(bodyContent, "\"")
	n.OutputString.Swap(&bodyContent)

	return nil
}
