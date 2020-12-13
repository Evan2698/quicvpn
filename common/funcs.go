package common

import (
	"net"
	"quicvpn/tun"
)

func CreateTun(name, ip, mask string, mtu int) (tun.Tun, error) {

	tdev := tun.NewTun()

	addr := net.ParseIP(ip)
	ma := net.ParseIP(mask)

	err := tdev.CreateTun(name, mtu, addr, ma)
	if err != nil {
		return nil, err
	}

	return tdev, nil
}
