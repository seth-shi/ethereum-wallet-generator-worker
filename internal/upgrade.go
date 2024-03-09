package internal

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	binNameMap = map[string]string{
		"windows": "ethereum-wallet-generator-nodes.exe",
		"linux":   "ethereum-wallet-generator-nodes",
		"darwin":  "ethereum-wallet-generator-nodes",
	}
	linkMap = map[string]map[string]string{
		"windows": {
			"amd64": "https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-windows-amd64.zip",
		},
		"linux": {
			"amd64": "https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-linux-amd64.tar.gz",
			"arm64": "https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-linux-arm64.tar.gz",
		},
		"darwin": {
			"amd64": "https://github.com/seth-shi/ethereum-wallet-generator-nodes/releases/download/v9.9.9/ethereum-wallet-generator-nodes-v9.9.9-darwin-amd64.tar.gz",
			"arm64": "ethereum-wallet-generator-nodes-v9.9.9-darwin-arm64.tar.gz\n",
		},
	}
)

type Upgrade struct {
}

func NewUpgrade() *Upgrade {
	return &Upgrade{}
}

func (u *Upgrade) Run() error {

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH

	archList, exists := linkMap[osName]
	if !exists {
		return errors.New(fmt.Sprintf("你的系统是[%s]仅支持[%s]", osName, strings.Join(lo.Keys(linkMap), ",")))
	}

	link, exists := archList[arch]
	if !exists {
		return errors.New(fmt.Sprintf("你的 CPU 架构是[%s]仅支持[%s]", arch, strings.Join(lo.Keys(archList), ",")))
	}

	packagePath := filepath.Join(pwd, path.Base(link))
	binPath := filepath.Join(pwd, binNameMap[osName])
	if err := deleteFileIfExists(packagePath); err != nil {
		return err
	}

	fmt.Printf("下载文件中[%s]->[%s]...\n", link, packagePath)
	if err := downloadFile(link, packagePath); err != nil {
		return err
	}

	// 解压文件
	fmt.Printf("删除老版本文件[%s]...\n", binPath)
	// 先删除可执行文件
	if err := u.deleteLocalExecuteFile(binPath); err != nil {
		return err
	}
	// 解压文件
	fmt.Printf("解压文件中[%s]->[%s]...\n", packagePath, binPath)
	if err := u.deCompressFile(packagePath, pwd); err != nil {
		return err
	}

	// 验证二进制文件是否存在
	fmt.Println("更新完成")

	return nil
}

func (u *Upgrade) deCompressFile(packagePath, outputPath string) error {

	// 在 windows 下直接删除文件会报错占用, 重命名不会
	if strings.HasSuffix(packagePath, ".zip") {
		if err := Unzip(packagePath, outputPath); err != nil {
			return err
		}
	} else {
		if err := UnTarGz(packagePath, outputPath); err != nil {
			return err
		}
	}

	return nil
}

func (u *Upgrade) deleteLocalExecuteFile(binPath string) error {

	// 文件不存在
	if _, err := os.Stat(binPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	// 在 windows 下直接删除文件会报错占用, 重命名不会
	return os.Rename(binPath, binPath+".bak")
}

func deleteFileIfExists(file string) error {

	// 文件不存在
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	// 删除文件
	return os.Remove(file)
}

type downloader struct {
	io.Reader       // 读取器
	Total     int64 // 总大小
	Current   int64 // 当前大小
}

func (d *downloader) Read(p []byte) (n int, err error) {
	n, err = d.Reader.Read(p)
	d.Current += int64(n)
	fmt.Printf("\r正在下载，下载进度：%.2f%%", float64(d.Current*10000/d.Total)/100)
	if d.Current == d.Total {
		fmt.Printf("\r下载完成，下载进度：%.2f%%\n", float64(d.Current*10000/d.Total)/100)
	}
	return
}

func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	myDownloader := &downloader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
	}
	if _, err := io.Copy(file, myDownloader); err != nil {
		return err
	}

	return nil
}
