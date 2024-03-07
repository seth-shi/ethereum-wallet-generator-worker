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
	TotalCount atomic.Uint64
	FoundCount atomic.Int32
	// 每次上报之后清空
	// 用以速度
	OnceCount atomic.Int32
	OnceUnix  atomic.Int64

	FilePoint *os.File

	StartAt time.Time

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
}

func NewNode(host string, cfg GetConfigRequest, c uint) (*Node, error) {

	// 打开或创建一个csv文件，以追加模式写入
	pf, err := os.OpenFile("wallet.node.csv", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Node{
		Host:       host,
		FilePoint:  pf,
		C:          c,
		Name:       generateNodeName(),
		Cip:        getCipher(cfg.Key),
		Config:     cfg,
		HttpClient: resty.New().SetTimeout(time.Second * 5),
	}, nil
}

func (n *Node) Run() {

	// 启动上报一次
	MustError(n.reportServer(nil))
	for i := 0; i < int(n.C); i++ {
		go n.run()
	}

	// 定时上报状态
	n.timerReportServer()
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

func (n *Node) run() {

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
		n.OnceCount.Add(1)
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

	now := time.Now().Unix()
	recentCount := n.OnceCount.Swap(0)
	lastUnix := n.OnceUnix.Swap(now)
	defer func() {
		if err != nil {
			// 恢复数量
			n.OnceCount.Add(recentCount)
			n.OnceUnix.Add(lastUnix)
		}
	}()

	// 计算时间
	var speed = 1.0
	if diffSeconds := float64(now - lastUnix); diffSeconds > 0 {
		speed = float64(recentCount) / diffSeconds
	}
	progressReq := &NodeProgress{
		Name:       n.Name,
		Count:      int(recentCount),
		Found:      int(n.FoundCount.Load()),
		Speed:      speed,
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

	tm.Clear()
	tm.MoveCursor(0, 0)
	// 永远返回不失败
	_, _ = tm.Println(fmt.Sprintf("节点名:%s 线程*%d 服务器:%s", n.Name, n.C, n.Host))
	_, _ = tm.Println(fmt.Sprintf("总生成:%d 总找到:%d", n.TotalCount.Load(), n.FoundCount.Load()))
	_, _ = tm.Println(bodyContent)
	tm.Flush()

	return nil
}
