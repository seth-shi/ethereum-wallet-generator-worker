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
	config       *GetConfigRequest
	buildVersion string
	// 运行配置
	port    int
	startAt time.Time

	// 无锁输出
	screenOutput string
	nodes        *orderedmap.OrderedMap[string, *NodeStatusRequest]
	locker       sync.RWMutex

	// 数据文件
	filePoint *os.File
	// 加密解密
	key []byte
	// 是否需要清屏
	needClearScreen bool
}

func NewMaster(port int, prefix, suffix, key string) (*Master, error) {

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

	key = strings.TrimSpace(key)
	if key == "" {
		key = lo.RandomString(16, lo.LowerCaseLettersCharset)
	}

	if len(key) != keyLength {
		return nil, errors.New("无效的秘钥,必须是16位")
	}

	master := &Master{
		port:         port,
		buildVersion: GetBuildVersion(),
		config: &GetConfigRequest{
			Prefix:   prefix,
			Suffix:   suffix,
			MayCount: uint64(math.Pow(16, float64(matchLength))),
		},
		filePoint:       walletPf,
		key:             []byte(key),
		nodes:           orderedmap.NewOrderedMap[string, *NodeStatusRequest](),
		startAt:         time.Now(),
		needClearScreen: true,
	}
	// 写入此次使用的 key
	if err := master.storeWalletData(key, "看仓库 readme 首页解密"); err != nil {
		return nil, err
	}

	return master, nil
}

func (m *Master) Stop() {

	MustError(m.filePoint.Close())
}

func (m *Master) Run() {

	go m.StartWebServer()

	ticker := time.NewTicker(time.Second * 1)

	tm.Flush()
	for range ticker.C {

		m.locker.Lock()
		nodes := m.nodes.Copy()
		m.locker.Unlock()

		m.output(nodes)
	}
}

func (m *Master) output(nodes *orderedmap.OrderedMap[string, *NodeStatusRequest]) {

	tableContent := m.buildContent(nodes)
	m.screenOutput = url.QueryEscape(tableContent)

	if m.needClearScreen {
		tm.Clear()
		m.needClearScreen = false
	}
	tm.MoveCursor(0, 0)
	_, _ = tm.Println(strings.Repeat("-", lineCharCount))
	_, _ = tm.Print(fmt.Sprintf("--版本号:%s\n--服务端:http://%s:%d?%s=%s\n", m.buildVersion, IPV4(), m.port, keyFieldName, m.key))
	_, _ = tm.Println(strings.Repeat("-", lineCharCount))
	_, _ = tm.Println(tableContent)
	tm.Flush()
}

func (m *Master) buildContent(renderNodes *orderedmap.OrderedMap[string, *NodeStatusRequest]) string {

	var (
		genCount    uint64
		walletCount uint64
		speed       float64
	)

	nowUnix := time.Now().Unix()
	data := lo.Map(renderNodes.Keys(), func(key string, i int) []string {
		item, _ := renderNodes.Get(key)
		activeUnix := item.LastActiveAt.Unix()

		// 虽然不活跃但是还是要计算总量
		genCount += uint64(item.Count)
		walletCount += uint64(item.Found)

		// 如果超过十五秒钟无响应, 那么不要计算生成速度
		runAt := nowUnix - item.StartAt
		if nowUnix-activeUnix > 15 {
			runAt = activeUnix - item.StartAt
			item.Speed = 0
			m.needClearScreen = true
		}
		speed += item.Speed

		versionDiff := "√"
		if item.BuildVersion != m.buildVersion {
			versionDiff = "×"
		}

		return []string{
			strconv.Itoa(i),
			item.Name,
			strconv.Itoa(item.Found),
			strconv.Itoa(item.Count),
			"",
			fmt.Sprintf("%.2f", item.Speed),
			timeToString(runAt),
			versionDiff + item.BuildVersion,
		}
	})
	runTime := int64(time.Now().Sub(m.startAt).Seconds())
	process := (float64(genCount) / float64(m.config.MayCount)) * 100

	tableBuf := &bytes.Buffer{}
	table := tablewriter.NewWriter(tableBuf)
	table.SetHeader([]string{"#", "节点", "已找到", "已生成", "", "速度", "运行时间", "版本号"})
	data = append(data, []string{
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
	})
	data = append(data, []string{
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
		"--------------",
	})
	data = append(data, []string{
		"运行时间",
		"预计时间",
		"总找到",
		"总生成",
		"",
		"",
		"前缀",
		"后缀",
	})

	data = append(data, []string{
		timeToString(runTime),
		timeToString(int64(float64(m.config.MayCount) / speed)),
		fmt.Sprintf("%d", walletCount),
		fmt.Sprintf("%d", genCount),
		"",
		"",
		m.config.Prefix,
		m.config.Suffix,
	})

	table.SetFooter([]string{
		"生成速度",
		fmt.Sprintf("%.2f 钱包/秒", speed),
		"预计要",
		fmt.Sprintf("%d", m.config.MayCount),
		"",
		"",
		"进度",
		fmt.Sprintf("%.2f%s", process, "%"),
	})
	table.AppendBulk(data)
	table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
	return tableBuf.String()
}

func (m *Master) updateNode(pro *NodeStatusRequest) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if oldPro, exists := m.nodes.Get(pro.Name); exists {
		pro.Count += oldPro.Count
	}
	pro.LastActiveAt = time.Now()
	m.nodes.Set(pro.Name, pro)
}
func (m *Master) StartWebServer() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {

		key, exists := c.GetQuery("key")
		if !exists || key == "" {
			c.JSON(http.StatusBadRequest, "key 不存在")
			return
		}

		if key != string(m.key) {
			c.JSON(http.StatusBadRequest, "和服务端的 key 不匹配")
			return
		}

		c.JSON(http.StatusOK, m.config)
	})
	// 上报状态
	r.POST("/", func(c *gin.Context) {

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		var pro NodeStatusRequest
		if err := json.Unmarshal(body, &pro); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// 写入成功数据
		m.updateNode(&pro)
		if pro.Address != nil && pro.EncryptMnemonic != nil {
			MustError(m.storeWalletData(lo.FromPtr(pro.Address), lo.FromPtr(pro.EncryptMnemonic)))
		}

		c.JSON(http.StatusOK, m.screenOutput)
	})

	addr := fmt.Sprintf(":%d", m.port)
	MustError(r.Run(addr))
}

func (m *Master) storeWalletData(address string, data string) error {

	// 创建一个csv写入器
	writer := csv.NewWriter(m.filePoint)
	// 循环写入数据
	err := writer.Write([]string{address, data})
	if err != nil {
		return errors.New(fmt.Sprintf("写入失败:[%s,%s]%s", address, data, err.Error()))
	}
	// 刷新缓冲区，确保所有数据都写入文件
	writer.Flush()
	return nil
}
