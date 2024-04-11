package master

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
)

type Master struct {
	matchConfig *models.MatchConfig
	runConfig   *RunConfig

	workerStatusManager *models.WorkerStatusManager
	// 无锁输出
	Title         string
	WorkerContent string
}

func NewMaster(port int, prefix, suffix string) (*Master, error) {

	var (
		cacheStatus = getStatusByCache()
		works       []*models.WorkStatusRequest
		startAt     = time.Now()
	)
	if cacheStatus != nil {
		fmt.Printf("进度从缓存中恢复:%s", cacheStatus.StartAt.Format(time.DateTime))
		startAt = cacheStatus.StartAt
		works = cacheStatus.Workers
	}

	rc, err := newRunConfig(port, startAt)
	if err != nil {
		return nil, err
	}

	master := &Master{
		Title: fmt.Sprintf(
			"--版本号:%s\n--服务端:http://%s:%d\n",
			rc.Version,
			utils.IPV4(),
			rc.Port,
		),
		matchConfig:         models.NewMatchConfig(prefix, suffix),
		runConfig:           rc,
		workerStatusManager: models.NewWorkerStatusManager(works),
	}

	if err := master.runConfig.storeWalletData(CsvHeaders); err != nil {
		return nil, err
	}

	return master, nil
}

func (m *Master) Run() error {

	go m.StartWebServer()
	go m.tickerSaveRunStatus()

	var lastMinute = time.Now().Minute()
	for ts := range time.Tick(time.Second) {

		nowMinute := ts.Minute()
		if nowMinute > lastMinute {
			lastMinute = nowMinute
			fmt.Print("\033[2J\033[H")
		}

		m.output(m.workerStatusManager.All())
	}

	return m.runConfig.FilePoint.Close()
}

func (m *Master) output(workers []*models.WorkStatusRequest) {

	fmt.Printf(
		"\u001B[H%s\n%s\n%s\n%s",
		strings.Repeat("-", consts.LineCharCount),
		m.Title,
		strings.Repeat("-", consts.LineCharCount),
		m.buildContent(workers),
	)
}

func (m *Master) buildContent(renderWorkers []*models.WorkStatusRequest) string {

	var (
		genCount    uint64
		walletCount uint64
		speed       float64
	)

	nowUnix := time.Now().Unix()
	genCount = lo.SumBy(renderWorkers, func(ww *models.WorkStatusRequest) uint64 {
		return uint64(ww.Count)
	})
	data := lo.Map(renderWorkers, func(item *models.WorkStatusRequest, i int) []string {
		activeUnix := item.LastActiveAt.Unix()

		// 虽然不活跃但是还是要计算总量
		walletCount += uint64(item.Found)

		// 如果超过十五秒钟无响应, 那么不要计算生成速度
		runAt := nowUnix - item.StartAt
		if nowUnix-activeUnix > 15 {
			runAt = activeUnix - item.StartAt
			item.Speed = 0
		}
		speed += item.Speed

		versionDiff := "√"
		if item.BuildVersion != m.runConfig.Version {
			versionDiff = "×"
		}

		var genProcess = 0.0
		if genCount > 0 {
			genProcess = float64(item.Count) / float64(genCount)
		}

		return []string{
			strconv.Itoa(i),
			item.Name,
			strconv.Itoa(item.Found),
			strconv.Itoa(item.Count),
			fmt.Sprintf("%.2f%s", genProcess*100, "%"),
			fmt.Sprintf("%.2f", item.Speed),
			utils.TimeToString(runAt),
			versionDiff + item.BuildVersion,
		}
	})
	runTime := int64(time.Now().Sub(m.runConfig.StartAt).Seconds())
	process := (float64(genCount) / float64(m.matchConfig.MayCount)) * 100
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
		utils.TimeToString(runTime),
		utils.TimeToString(int64(float64(m.matchConfig.MayCount) / speed)),
		fmt.Sprintf("%d", walletCount),
		fmt.Sprintf("%d", genCount),
		"",
		"",
		m.matchConfig.Prefix,
		m.matchConfig.Suffix,
	})
	footer := []string{
		"生成速度",
		fmt.Sprintf("%.2f 钱包/秒", speed),
		"预计要",
		fmt.Sprintf("%d", m.matchConfig.MayCount),
		"",
		"",
		"进度",
		fmt.Sprintf("%.2f%s", process, "%"),
	}

	buf := &strings.Builder{}
	tab := tablewriter.NewWriter(buf)
	tab.SetHeader([]string{"#", "节点", "已找到", "已生成", "占比", "速度", "运行时间", "版本号"})
	tab.AppendBulk(data)
	tab.SetFooter(footer)
	tab.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	tab.SetAlignment(tablewriter.ALIGN_LEFT)
	tab.Render()

	originContent := buf.String()
	m.WorkerContent = url.QueryEscape(originContent)
	return originContent
}

func (m *Master) StartWebServer() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, m.matchConfig)
	})
	// 上报状态
	r.POST("/", func(c *gin.Context) {

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		var pro models.WorkStatusRequest
		if err := json.Unmarshal(body, &pro); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// 写入成功数据
		m.workerStatusManager.Add(&pro)
		if pro.HasWallet() {
			// 如果没有秘钥, 那么就是客户端发的
			utils.MustError(m.runConfig.storeWalletData(pro.Wallet))
		}

		c.JSON(http.StatusOK, m.WorkerContent)
	})

	addr := fmt.Sprintf(":%d", m.runConfig.Port)
	utils.MustError(r.Run(addr))
}

func (m *Master) tickerSaveRunStatus() {
	for range time.Tick(time.Minute * 1) {

		data := models.MasterRunStatusCache{
			Workers: m.workerStatusManager.All(),
			StartAt: m.runConfig.StartAt,
		}
		utils.ShowIfError(setStatusToCache(data))
	}
}
