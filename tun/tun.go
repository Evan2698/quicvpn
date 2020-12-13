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
	CreateTun(name string, mtu int, IP, mask net.IP) error
	Read() ([]byte, error)
	Write([]byte) (int, error)
	Close() error
}

func NewTun() Tun {

	return &tundevice{}
}
