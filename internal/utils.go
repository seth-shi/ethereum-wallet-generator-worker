package internal

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/golang-module/dongle"
	"log"
	"math"
	"net"
	"os/user"
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

func getNodeName() string {

	var name string
	var address string

	if netInterfaces, err := net.Interfaces(); err == nil {
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) != 0 {
				address = strings.ReplaceAll(strings.TrimSpace(macAddr), ":", "")
				break
			}
		}
	}

	if u, err := user.Current(); err == nil {
		name = u.Name
		if address == "" {
			address = strings.ReplaceAll(u.Uid, "-", "")
		}
	}

	address = fixLength(address, 16)

	return fmt.Sprintf("%s@%s", name, address)
}

func fixLength(str string, length int) string {

	if len(str) >= length {
		return str[:length]
	}

	return str + strings.Repeat("0", length-len(str))
}
