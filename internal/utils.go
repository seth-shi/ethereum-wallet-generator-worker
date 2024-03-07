package internal

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/golang-module/dongle"
	"log"
	"math"
	"net"
	"os"
	"os/user"
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

	if SubTime == 0 {
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

func generateNodeName() string {

	var name string
	var address string
	if host, err := os.Hostname(); err == nil {
		name = host
	}

	if name == "" {
		if u, err := user.Current(); err == nil {
			name = u.Name
		}
	}

	if netInterfaces, err := net.Interfaces(); err == nil {
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) != 0 {
				address = macAddr
			}
		}
	}

	if address == "" {
		address = getIpName()
	}

	return fmt.Sprintf("%s@%s", name, address)
}

func getIpName() string {

	if interfaceAddr, err := net.InterfaceAddrs(); err == nil {
		addrLastIndex := len(interfaceAddr) - 1
		for i, addr := range interfaceAddr {

			if addrLastIndex == i {
				return addr.String()
			}

			ipNet, isValidIpNet := addr.(*net.IPNet)
			if isValidIpNet && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}

	return IPV4()
}
