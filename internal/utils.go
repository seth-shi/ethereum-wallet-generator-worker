package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"log"
	"math"
	"net"
	"os/user"
	"runtime/debug"
	"strings"
	"time"
)

func MustError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func GetBuildVersion() string {

	var (
		buildTime string
	)

	if buildInfo, ok := debug.ReadBuildInfo(); ok {

		for _, s := range buildInfo.Settings {
			switch s.Key {
			//case "vcs.revision":
			//	git = s.Value
			case "vcs.time":
				buildTime = s.Value
			}
		}
	}

	// 如果不为空
	if buildTime != "" {
		if timeObj, err := time.Parse(time.RFC3339Nano, buildTime); err == nil {
			buildTime = timeObj.Format("20060102150405")
		}
	}

	return buildTime
}

func AesGcmEncrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	// Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func AesGcmDecrypt(ciphertext string, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Base64解码
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertextBytes := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertextBytes, nil)
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
