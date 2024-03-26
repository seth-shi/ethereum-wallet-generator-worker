package main

import (
	"errors"
	"fmt"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/master"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/worker"
	"github.com/urfave/cli/v2"
	"net/http"
	"runtime"
	"strings"
)

const (
	mnemonicCount = 12
)

var (
	masterCommand = &cli.Command{
		Name:  "master",
		Usage: "启动 HTTP 服务器, 收集信息",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Value: 8000,
				Usage: "HTTP 服务端口",
			},

			&cli.StringFlag{
				Name:  "prefix",
				Value: "",
				Usage: "钱包地址前缀",
			},
			&cli.StringFlag{
				Name:  "suffix",
				Value: "",
				Usage: "钱包地址后缀",
			},
			&cli.StringFlag{
				Name:  "key",
				Value: "",
				Usage: "通讯 key,如果不指定,那么自动生成",
			},
		},
		Before: func(cCtx *cli.Context) (err error) {

			var (
				prefix = cCtx.String("prefix")
				suffix = cCtx.String("suffix")
				key    = cCtx.String("key")
				port   = cCtx.Int("port")
			)
			if cCtx.String("prefix") == "" && cCtx.String("suffix") == "" {
				return errors.New("钱包前缀和后缀不能同时为空")
			}

			if Master, err = master.NewMaster(port, key, prefix, suffix); err != nil {
				return
			}

			return nil
		},
		Action: func(cCtx *cli.Context) error {

			return Master.Run()
		},
	}
	workerCommand = &cli.Command{
		Name:  "worker",
		Usage: "生成节点",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "server",
				Value: "http://localhost:8000",
				Usage: "服务器地址",
			},
			&cli.UintFlag{
				Name:  "c",
				Value: 0,
				Usage: "并发线程, 默认 CPU 个数",
			},
			&cli.StringFlag{
				Name:  "name",
				Value: "",
				Usage: "自定义节点名字",
			},
		},
		Before: func(cCtx *cli.Context) error {
			var (
				c          = cCtx.Uint("c")
				serverHost = cCtx.String("server")
				workerName = cCtx.String("name")
			)
			if c == 0 {
				c = uint(runtime.NumCPU())
			}

			if workerName == "" {
				workerName = utils.GenWorkerName()
			}
			fmt.Printf("worker:[%s]启动中...\n", workerName)

			// 从服务端获取配置
			mc, resp, err := utils.GetMatchConfig(serverHost)
			if err != nil {
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.New(fmt.Sprintf("获取配置失败[%d]%s", resp.StatusCode(), resp.String()))
			}

			Worker, err = worker.NewWorker(serverHost, mc, c, workerName)
			return err
		},
		Action: func(cCtx *cli.Context) error {

			return Worker.Run()
		},
	}
	decryptCommand = &cli.Command{
		Name:  "decrypt",
		Usage: "解密钱包数据",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "data",
				Value: "",
				Usage: "加密的数据",
			},
			&cli.StringFlag{
				Name:  "key",
				Value: "",
				Usage: "服务器运行的秘钥",
			},
			&cli.UintFlag{
				Name:  "offset",
				Value: 0,
				Usage: "偏移量",
			},
			&cli.UintFlag{
				Name:  "limit",
				Value: 12,
				Usage: "输出词的数量",
			},
		},
		Action: func(cCtx *cli.Context) error {

			var (
				key    = cCtx.String("key")
				data   = cCtx.String("data")
				offset = cCtx.Uint("offset")
				limit  = cCtx.Uint("limit")
			)

			if key == "" {
				return errors.New("秘钥不能为空")
			}
			if data == "" {
				return errors.New("加密数据不能为空")
			}
			count := offset + limit
			if count > mnemonicCount {
				return errors.New("助记词只能返回12个")
			}

			decryptBytes, err := utils.AesGcmDecrypt(data, []byte(key))
			if err != nil {
				return errors.New(fmt.Sprintf("解密失败:%s", err.Error()))
			}

			decryptData := strings.Split(string(decryptBytes), " ")
			if len(decryptData) != mnemonicCount {
				return errors.New(fmt.Sprintf("助记词个数不正确:[%s]", decryptBytes))
			}

			end := limit + offset
			fmt.Printf("助记词 %d-%d 开始\n", offset, end)
			for i := offset; i < end; i++ {
				fmt.Printf("%s ", decryptData[i])
			}
			fmt.Printf("\n助记词 %d-%d 结束\n", offset, end)

			return nil
		},
	}
)
