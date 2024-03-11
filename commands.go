package main

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal"
	"github.com/urfave/cli/v2"
	"net/http"
	"runtime"
	"strings"
	"time"
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

			if Master, err = internal.NewMaster(port, prefix, suffix, key); err != nil {
				return
			}

			return nil
		},
		Action: func(cCtx *cli.Context) error {

			defer func() {
				Master.FilePoint.Close()
			}()

			Master.Run()

			return nil
		},
	}
	nodeCommand = &cli.Command{
		Name:  "node",
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
				nodeName   = cCtx.String("name")
			)
			if c == 0 {
				c = uint(runtime.NumCPU())
			}

			if nodeName == "" {
				nodeName = internal.GetNodeName()
			}
			fmt.Printf("节点:[%s]启动中...\n", nodeName)

			// 从服务端获取配置
			var apiRes internal.GetConfigRequest
			resp, err := resty.New().SetTimeout(time.Second * 3).R().SetResult(&apiRes).Get(serverHost)
			if err != nil {
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.New(fmt.Sprintf("获取配置失败[%d]%s", resp.StatusCode(), resp.String()))
			}

			if Node, err = internal.NewNode(serverHost, apiRes, c, nodeName); err != nil {
				return err
			}

			return nil
		},
		Action: func(cCtx *cli.Context) error {

			Node.Run()
			return nil
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

			decryptBytes, err := internal.AesGcmDecrypt(data, []byte(key))
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
