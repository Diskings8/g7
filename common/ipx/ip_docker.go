package ipx

import (
	"net"
	"strings"
)

func GetContainerIP() string {
	// 遍历所有网卡
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range ifaces {
		// 跳过无效网卡
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取IP地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			ip := ipNet.IP.To4()
			if ip != nil && strings.HasPrefix(ip.String(), "172.") { // Docker 内网都是 172 开头
				return ip.String()
			}
		}
	}

	return "127.0.0.1"
}
