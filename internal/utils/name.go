package utils

import (
	"fmt"
	"net"
	"os/user"
	"runtime/debug"
	"strings"
	"time"
)

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

func GenWorkerName() string {

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
		prefix = u.Name
		if value == "" {
			value = u.Uid
			suffix = "uid"
		}
	}

	return fmt.Sprintf("%s@%s-%s", prefix, suffix, value)
}
