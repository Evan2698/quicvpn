package tun

import (
	"net"
	"os"
)

type tundevice struct {
	mtu  int
	fd   *os.File
	name string
}

type Tun interface {
	OpenTun(addr net.IP, network net.IP, mask net.IP) error
	Read() ([]byte, error)
	Write([]byte) (int, error)
	Close() error
}

func NewTun(MTU int) Tun {

	return &tundevice{
		mtu: MTU,
	}
}