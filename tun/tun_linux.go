package tun

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	IFF_NO_PI = 0x10
	IFF_TUN   = 0x01
	IFF_TAP   = 0x02
	TUNSETIFF = 0x400454CA
)

const (
	IPv6_HEADER_LENGTH = 40
)

func (d *tundevice) OpenTun(addr net.IP, network net.IP, mask net.IP) error {

	deviceFile := "/dev/net/tun"
	fd, err := os.OpenFile(deviceFile, os.O_RDWR, 0)
	if err != nil {
		log.Println("[CRIT] Note: Cannot open TUN/TAP dev", deviceFile, err)
		return err
	}
	d.fd = fd

	ifr := make([]byte, 18)
	ifr[17] = IFF_NO_PI
	ifr[16] = IFF_TUN
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(d.fd.Fd()), uintptr(TUNSETIFF),
		uintptr(unsafe.Pointer(&ifr[0])))
	if errno != 0 {
		log.Println("[CRIT] Cannot ioctl TUNSETIFF:", errno)
	}

	d.name = string(ifr)
	d.name = d.name[:strings.Index(d.name, "\000")]
	log.Printf("[INFO] TUN/TAP device %s opened.", d.name)
	if err := d.setupAddress(addr.String(), mask.String()); err != nil {
		return err
	}

	return nil
}

func (d *tundevice) setupAddress(addr, mask string) error {
	cmd := exec.Command("ifconfig", d.name, addr,
		"netmask", mask, "mtu", strconv.Itoa(d.mtu))
	log.Printf("[DEBG] ifconfig command: %v", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		log.Printf("[EROR] Linux ifconfig failed: %v.", err)
		return err
	}
	return nil
}

func (d *tundevice) Write(in []byte) (int, error) {
	return d.fd.Write(in)
}

func (d *tundevice) Read() ([]byte, error) {
	buf := make([]byte, d.mtu)
	n, err := d.fd.Read(buf)
	if err != nil {
		log.Println("read tun device failed", err)
		return nil, err
	}

	totalLen := 0
	switch buf[0] & 0xf0 {
	case 0x40:
		totalLen = 256*int(buf[2]) + int(buf[3])
	case 0x60:
		totalLen = 256*int(buf[4]) + int(buf[5]) + IPv6_HEADER_LENGTH
	}
	if totalLen != n && totalLen <= d.mtu {
		return nil, fmt.Errorf("read n(%v)!=total(%v)", n, totalLen)
	}
	return buf[:n], nil
}

func (d *tundevice) Close() error {
	if d.fd != nil {
		d.fd.Close()
		d.fd = nil
	}
	return nil
}
