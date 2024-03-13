package internal

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	tm "github.com/buger/goterm"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Node struct {
	name string

	buildVersion string

	// 总量
	totalCount  atomic.Uint64
	foundCount  atomic.Int32
	recentCount atomic.Int32
	startAt     int64

	// 数据保存
	filePoint *os.File

	// 服务端地址
	host string

	// 线程数量
	c uint

	// 加密解密
	key []byte

	// 配置
	config GetConfigRequest

	// 请求
	httpClient *resty.Client

	outputString atomic.Pointer[string]
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
		host:         host,
		filePoint:    pf,
		c:            c,
		name:         nodeName,
		key:          []byte(key),
		config:       cfg,
		buildVersion: GetBuildVersion(),
		startAt:      time.Now().Unix(),
		httpClient:   resty.New().SetTimeout(time.Second * 5),
	}, nil
}

func (n *Node) Stop() {
	MustError(n.filePoint.Close())
}

func (n *Node) Run() error {

	ctx, cancel := NewSignal()
	defer cancel()

	// 启动上报一次
	MustError(n.reportServer(nil))
	for i := 0; i < int(n.c); i++ {
		go n.loopMatchWallets(ctx)
	}

	// 定时上报状态
	go n.timerReportServer(ctx)

	// 刷新输出
	go n.timerOutput(ctx)

	<-ctx.Done()

	// 信号中止的时候, 再输出一次
	tm.Clear()
	n.output(time.Now().Unix())
	n.Stop()

	time.Sleep(time.Second * 3)
	tm.Println("程序停止")

	return nil
}
func (n *Node) speed(nowUnix int64) float64 {
	var speed = 0.0
	diff := nowUnix - n.startAt
	if diff > 0 {
		speed = float64(n.totalCount.Load()) / float64(diff)
	}
	return speed
}
func (n *Node) timerOutput(ctx context.Context) {
	timer := time.NewTicker(time.Second)
	tm.Clear()
	var lastMinute = time.Now().Minute()
	for {

		select {
		case <-ctx.Done():
			timer.Stop()
			fmt.Println("定时输出停止")
			return
		case ts := <-timer.C:
			nowMinute := ts.Minute()
			if nowMinute > lastMinute {
				lastMinute = nowMinute
				tm.Clear()
			}

			n.output(ts.Unix())
		}
	}
}

func (n *Node) output(unix int64) {
	// 永远返回不失败
	tm.MoveCursor(0, 0)
	_, _ = tm.Println(strings.Repeat("-", lineCharCount))
	tm.MoveCursor(0, 2)
	_, _ = tm.Println(fmt.Sprintf("--版本号:%s", n.buildVersion))
	_, _ = tm.Println(fmt.Sprintf("--节点名:%s 线程*%d", n.name, n.c))
	_, _ = tm.Println(fmt.Sprintf(
		"--实时速度: %.2f 钱包/秒 生成:%d 找到:%d",
		n.speed(unix),
		n.totalCount.Load(),
		n.foundCount.Load(),
	))
	_, _ = tm.Println(*n.outputString.Load())
	tm.Flush()
}

func (n *Node) timerReportServer(ctx context.Context) {

	// 上报时长 5s
	timer := time.NewTicker(time.Second * 5)
	for {

		select {
		case <-ctx.Done():
			timer.Stop()
			fmt.Println("定时上报停止")
			return
		case <-timer.C:
			if err := n.reportServer(nil); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (n *Node) loopMatchWallets(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			fmt.Println("查找钱包停止")
			return
		default:
			if newWalletData := n.matchNewWallet(); newWalletData != nil {
				if err := n.reportServer(newWalletData); err != nil {
					if storeErr := n.storeWalletData(newWalletData); storeErr != nil {
						MustError(storeErr)
					}
				}
			}
		}
	}
}

func (n *Node) storeWalletData(wa *Wallet) error {

	// 凡是出错, 直接打印原始出来在标准输出
	// node 保存钱包的时候, 也需要加密数据
	encryptData, err := AesGcmEncrypt(wa.Mnemonic, n.key)
	if err != nil {
		return errors.New(fmt.Sprintf("钱包加密失败:[%s,%s]%s", wa.Address, wa.Mnemonic, err.Error()))
	}

	// 创建一个csv写入器
	writer := csv.NewWriter(n.filePoint)
	// 循环写入数据
	if err := writer.Write([]string{wa.Address, encryptData}); err != nil {
		return errors.New(fmt.Sprintf("钱包写入失败:[%s,%s]%s", wa.Address, wa.Mnemonic, err.Error()))
	}

	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
	return nil
}

func (n *Node) matchNewWallet() *Wallet {
	defer func() {
		n.totalCount.Add(1)
		n.recentCount.Add(1)
	}()

	wallet, err := newWallet()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if n.config.Prefix != "" && !strings.HasPrefix(wallet.Address, n.config.Prefix) {
		return nil
	}

	if n.config.Suffix != "" && !strings.HasSuffix(wallet.Address, n.config.Suffix) {
		return nil
	}

	n.foundCount.Add(1)

	return wallet
}

func (n *Node) reportServer(wa *Wallet) (err error) {

	nowUnix := time.Now().Unix()
	recentCount := n.recentCount.Swap(0)
	defer func() {
		if err != nil {
			// 恢复数量, 中途可能数量增加了
			n.recentCount.Add(recentCount)
		}
	}()

	// 计算时间
	progressReq := &NodeStatusRequest{
		Name:         n.name,
		BuildVersion: n.buildVersion,
		Count:        int(recentCount),
		Found:        int(n.foundCount.Load()),
		Speed:        n.speed(nowUnix),
		StartAt:      n.startAt,
	}
	if wa != nil {
		encryptData, err := AesGcmEncrypt(wa.Mnemonic, n.key)
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

	resp, err := n.httpClient.R().SetBody(data).Post(n.host)
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
	n.outputString.Swap(&bodyContent)

	return nil
}
