package internal

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/golang-module/dongle"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func MustError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func getCipher(key string) *dongle.Cipher {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.ECB)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(key)
	cipher.SetIV(key)
	return cipher
}

func IPV4() string {

	resp, err := resty.New().SetTimeout(time.Second * 3).R().Get("https://api.ipify.org")
	if err != nil {
		return ""
	}

	return resp.String()
}

func timeToString(SubTime int64) string {

	if SubTime <= 0 {
		return ""
	}

	// 秒
	if SubTime < 60 {
		return fmt.Sprintf("%d秒", SubTime)
	}

	// 分钟
	if SubTime < 60*60 {
		minute := int(math.Floor(float64(SubTime / 60)))
		second := SubTime % 60
		return fmt.Sprintf("%d分%d秒", minute, second)
	}

	// 小时
	if SubTime < 60*60*24 {
		hour := int(math.Floor(float64(SubTime / (60 * 60))))
		tail := SubTime % (60 * 60)
		minute := int(math.Floor(float64(tail / 60)))
		second := tail % 60
		return fmt.Sprintf("%d时%d分%d秒", hour, minute, second)
	}

	// 天
	day := int(math.Floor(float64(SubTime / (60 * 60 * 24))))
	tail := SubTime % (60 * 60 * 24)
	hour := int(math.Floor(float64(tail / (60 * 60))))
	tail = SubTime % (60 * 60)
	minute := int(math.Floor(float64(tail / 60)))
	second := tail % 60
	return fmt.Sprintf("%d天%d时%d分%d秒", day, hour, minute, second)
}

func GetNodeName() string {

	var prefix string
	var value string
	var suffix string

	if netInterfaces, err := net.Interfaces(); err == nil {
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) != 0 {
				value = strings.TrimSpace(macAddr)
				suffix = "mac"
				break
			}
		}
	}

	if u, err := user.Current(); err == nil {
		prefix = u.Username
		if value == "" {
			value = u.Uid
			suffix = "uid"
		}
	}

	return fmt.Sprintf("%s@%s-%s", prefix, suffix, value)
}

func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()
	var decodeName string
	for _, f := range zipReader.File {
		if f.Flags == 0 {
			//如果标致位是0  则是默认的本地编码   默认为gbk
			i := bytes.NewReader([]byte(f.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := io.ReadAll(decoder)
			decodeName = string(content)
		} else {
			//如果标志为是 1 << 11也就是 2048  则是utf-8编码
			decodeName = f.Name
		}
		fpath := filepath.Join(destDir, decodeName)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UnTarGz 解压缩 tar.gz 到指定文件夹
func UnTarGz(tarGzFile string, destDir string) error {
	gzipStream, err := os.Open(tarGzFile)
	if err != nil {
		return err
	}
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(uncompressedStream)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			outputPath := filepath.Join(destDir, header.Name)
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outputPath := filepath.Join(destDir, header.Name)
			outFile, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			outFile.Close()
		default:
			return err
		}
	}
	return nil
}
