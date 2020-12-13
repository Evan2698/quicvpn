package tun

import (
	"os"
)

type tundevice struct {
	mtu  int
	fd   *os.File
	name string
}

type Tun interface {
	Read() ([]byte, error)
	Write([]byte) (int, error)
	Close() error
}

func NewTun() Tun {

	return &tundevice{}
}
