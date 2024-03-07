package internal

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	tm "github.com/buger/goterm"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/dongle"
	"github.com/olekukonko/tablewriter"
	"github.com/samber/lo"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Master struct {
	Config *GetConfigRequest
	// 运行配置
	Port         int
	ServerPublic string
	StartAt      time.Time

	// 无锁输出
	ScreenOutput string
	Nodes        *orderedmap.OrderedMap[string, *NodeProgress]
	Locker       sync.RWMutex

	// 数据文件
	FilePoint *os.File
	// 加密解密
	Cip *dongle.Cipher
}

func NewMaster(port int, prefix, suffix string) (*Master, error) {

	// 打开或创建一个csv文件，以追加模式写入
	walletPf, err := os.OpenFile("wallet.csv", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	matchLength := len(suffix)
	if prefix != "" {
		matchLength += len(prefix)
		if !strings.HasPrefix(prefix, addressPrefix) {
			prefix = fmt.Sprintf("%s%s", addressPrefix, prefix)
		} else {
			matchLength -= len(addressPrefix)
		}
	}

	key := lo.RandomString(16, lo.LowerCaseLettersCharset)
	return &Master{
		Port: port,
		Config: &GetConfigRequest{
			Prefix:   prefix,
			Suffix:   suffix,
			Key:      key,
			MayCount: uint64(math.Pow(16, float64(matchLength))),
		},
		FilePoint:    walletPf,
		ServerPublic: fmt.Sprintf("服务端:http://%s:%d\n", IPV4(), port),
		Cip:          getCipher(key),
		Nodes:        orderedmap.NewOrderedMap[string, *NodeProgress](),
		StartAt:      time.Now(),
	}, nil
}

func (m *Master) Run() {

	ticker := time.NewTicker(time.Second * 1)
	for range ticker.C {

		m.Locker.Lock()
		nodes := m.Nodes.Copy()
		m.Locker.Unlock()

		m.output(nodes)
	}
}

func (m *Master) output(nodes *orderedmap.OrderedMap[string, *NodeProgress]) {

	tableContent := m.buildContent(nodes)
	m.ScreenOutput = url.QueryEscape(tableContent)

	tm.Clear()
	tm.MoveCursor(0, 0)
	// 永远返回不失败
	_, _ = tm.Println(m.ServerPublic)
	_, _ = tm.Println(tableContent)
	tm.Flush()
}

func (m *Master) buildContent(renderNodes *orderedmap.OrderedMap[string, *NodeProgress]) string {

	var (
		genCount    uint64
		walletCount uint64
		speed       float64
	)

	data := lo.Map(renderNodes.Keys(), func(key string, i int) []string {
		item, _ := renderNodes.Get(key)
		genCount += uint64(item.Count)
		walletCount += uint64(item.Found)
		speed += item.Speed
		return []string{
			strconv.Itoa(i),
			item.Name,
			strconv.Itoa(item.Count),
			strconv.Itoa(item.Found),
			fmt.Sprintf("%.2f 钱包/秒", item.Speed),
		}
	})
	runTime := int64(time.Now().Sub(m.StartAt).Seconds())
	process := (float64(genCount) / float64(m.Config.MayCount)) * 100

	tableBuf := &bytes.Buffer{}
	table := tablewriter.NewWriter(tableBuf)
	table.SetHeader([]string{"#", "节点", "已生成", "已找到", "速度"})
	data = append(data, []string{
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
	})
	data = append(data, []string{
		"运行时间",
		"预计时间",
		"",
		"总生成/预计要",
		"总找到/进度",
	})
	data = append(data, []string{
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
	})
	t := int64(float64(m.Config.MayCount) / speed)
	if t <= 0 {
		t = 0
	}
	data = append(data, []string{
		timeToString(runTime),
		timeToString(t),
		"",
		fmt.Sprintf("%d", genCount),
		fmt.Sprintf("%d", walletCount),
	})

	table.SetFooter([]string{
		"生成速度",
		fmt.Sprintf("%.2f 钱包/秒", speed),
		"",
		fmt.Sprintf("%d", m.Config.MayCount),
		fmt.Sprintf("%.2f%s", process, "%"),
	})
	table.AppendBulk(data)
	table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
	return tableBuf.String()
}

func (m *Master) updateNode(pro *NodeProgress) {
	m.Locker.Lock()
	defer m.Locker.Unlock()
	if oldPro, exists := m.Nodes.Get(pro.Name); exists {
		pro.Count += oldPro.Count
	}
	m.Nodes.Set(pro.Name, pro)
}
func (m *Master) StartWebServer() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, m.Config)
	})
	// 上报状态
	r.POST("/", func(c *gin.Context) {

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		data := dongle.Decrypt.FromRawBytes(body).ByAes(m.Cip).ToBytes()
		var pro NodeProgress
		if err := json.Unmarshal(data, &pro); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// 写入成功数据
		m.updateNode(&pro)
		if pro.WalletData != nil {
			MustError(m.storeWalletData(pro.WalletData))
		}

		c.JSON(http.StatusOK, m.ScreenOutput)
	})

	addr := fmt.Sprintf(":%d", m.Port)
	MustError(r.Run(addr))
}

func (m *Master) storeWalletData(data *Wallet) error {

	// 创建一个csv写入器
	writer := csv.NewWriter(m.FilePoint)
	// 循环写入数据
	err := writer.Write([]string{data.Address, data.Mnemonic})
	if err != nil {
		return errors.New(fmt.Sprintf("钱包写入失败:[%s,%s]%s", data.Address, data.Mnemonic, err.Error()))
	}
	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
	return nil
}
