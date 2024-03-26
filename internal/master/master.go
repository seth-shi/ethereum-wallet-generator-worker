package master

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/pprof"
	"io"
	"net/http"
	"strconv"
	"time"

	tm "github.com/buger/goterm"
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
	screenBuilder *ScreenBuilder
}

func NewMaster(port int, key, prefix, suffix string) (*Master, error) {

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

	rc, err := newRunConfig(port, key, startAt)
	if err != nil {
		return nil, err
	}

	master := &Master{
		matchConfig:         models.NewMatchConfig(prefix, suffix),
		runConfig:           rc,
		workerStatusManager: models.NewNodeStatusManager(works),
		screenBuilder:       newScreenBuilder(),
	}
	// 写入此次使用的 key
	if err := master.runConfig.storeWalletData(rc.key, "看仓库 readme 首页解密"); err != nil {
		return nil, err
	}

	return master, nil
}

func (m *Master) Run() error {

	go m.StartWebServer()
	go m.tickerSaveRunStatus()

	ticker := time.NewTicker(time.Second * 1)

	tm.Flush()
	for range ticker.C {
		m.output(m.workerStatusManager.All())
	}

	return m.runConfig.FilePoint.Close()
}

func (m *Master) output(workers []*models.WorkStatusRequest) {

	m.buildContent(workers)
	//if m.screenBuilder.NeedClearScreen {
	//	tm.Clear()
	//	m.screenBuilder.NeedClearScreen = false
	//}
	//tm.MoveCursor(0, 0)
	//_, _ = tm.Println(strings.Repeat("-", consts.LineCharCount))
	//_, _ = tm.Print(fmt.Sprintf(
	//	"--版本号:%s\n--服务端:http://%s:%d?%s=%s\n",
	//	m.runConfig.Version,
	//	utils.IPV4(),
	//	m.runConfig.Port,
	//	consts.QueryKeyFieldName,
	//	m.runConfig.key,
	//))
	//_, _ = tm.Println(strings.Repeat("-", consts.LineCharCount))
	//_, _ = tm.Println(m.screenBuilder.GetContent())
	//tm.Flush()
}

func (m *Master) buildContent(renderWorkers []*models.WorkStatusRequest) {

	var (
		genCount    uint64
		walletCount uint64
		speed       float64
	)

	nowUnix := time.Now().Unix()
	genCount = lo.SumBy(renderWorkers, func(node *models.WorkStatusRequest) uint64 {
		return uint64(node.Count)
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
			m.screenBuilder.NeedClearScreen = true
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

	// 生成 string
	m.screenBuilder.Build(data, footer)
}

func (m *Master) StartWebServer() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	pprof.Register(r)
	r.GET("/", func(c *gin.Context) {

		key, exists := c.GetQuery("key")
		if !exists || key == "" {
			c.JSON(http.StatusBadRequest, "key 不存在")
			return
		}

		if key != m.runConfig.key {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("秘钥[%s]和服务端的不匹配", key))
			return
		}

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
		if pro.Address != nil && pro.EncryptMnemonic != nil {
			utils.MustError(m.runConfig.storeWalletData(
				lo.FromPtr(pro.Address),
				lo.FromPtr(pro.EncryptMnemonic),
			))
		}

		c.JSON(http.StatusOK, m.screenBuilder.GetEncodeContent())
	})

	addr := fmt.Sprintf(":%d", m.runConfig.Port)
	utils.MustError(r.Run(addr))
}

func (m *Master) tickerSaveRunStatus() {
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		data := models.MasterRunStatusCache{
			Workers: m.workerStatusManager.All(),
			StartAt: m.runConfig.StartAt,
		}
		utils.ShowIfError(setStatusToCache(data))
	}
}
