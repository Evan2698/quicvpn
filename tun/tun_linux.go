package tun

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	IFF_NO_PI = 0x10
	IFF_TUN   = 0x01
	IFF_TAP   = 0x02
	TUNSETIFF = 0x400454CA
)

const (
	IPv6_HEADER_LENGTH = 40
	ifReqSize          = 80
)

func (d *tundevice) CreateTun(name string, mtu int, IP, mask net.IP) error {
	deviceFile := "/dev/net/tun"
	fd, err := os.OpenFile(deviceFile, os.O_RDWR, 0)
	if err != nil {
		log.Println("[CRIT] Note: Cannot open TUN/TAP dev", deviceFile, err)
		return err
	}

	d.fd = fd

	if len(name) > 14 {
		name = name[:14]
	}
	d.name = name
	ifr := make([]byte, 18)
	copy(ifr[:15], []byte(d.name))
	ifr[17] = IFF_NO_PI
	ifr[16] = IFF_TUN

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(d.fd.Fd()), uintptr(TUNSETIFF),
		uintptr(unsafe.Pointer(&ifr[0])))
	if errno != 0 {
		log.Println("[CRIT] Cannot ioctl TUNSETIFF:", errno)
		fd.Close()
		return errors.New("syscall  SYS_IOCTL failed")
	}

	d.name = string(ifr)
	d.name = d.name[:strings.Index(d.name, "\000")]
	log.Printf("[INFO] TUN/TAP device %s opened.", d.name)
	if err = configTun(d.name, IP, mask, d.mtu); err != nil {
		log.Println("config tun failed!", err)
		return errors.New("config tun failed!")
	}

	return nil

}

type socketAddrRequest struct {
	name [unix.IFNAMSIZ]byte
	addr unix.RawSockaddrInet4
}

type socketFlagsRequest struct {
	name  [unix.IFNAMSIZ]byte
	flags uint16
	pad   [22]byte
}

func setMTU(name string, fd, mtu int) error {
	var ifr [ifReqSize]byte
	copy(ifr[:], name)
	*(*uint32)(unsafe.Pointer(&ifr[unix.IFNAMSIZ])) = uint32(mtu)
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCSIFMTU),
		uintptr(unsafe.Pointer(&ifr[0])),
	)
	if errno < 0 {
		log.Println("set MTU failed,", errno)
		return errors.New("set mtu failed")
	}
	return nil
}

func setTunFlags(name string, fd int, set, clr uint16) error {

	var a socketFlagsRequest
	copy(a.name[:], name)
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCGIFFLAGS),
		uintptr(unsafe.Pointer(&a)),
	)

	if errno < 0 {
		log.Println("get flags failed", errno)
		return errors.New("setUp(get) failed")
	}

	a.flags = (a.flags & (^clr)) | set
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCSIFFLAGS),
		uintptr(unsafe.Pointer(&a)),
	)
	if errno < 0 {
		log.Println("set flags failed", errno)
		return errors.New("setUp(set) failed")
	}
	return nil

}

func setTunIP(name string, fd int, ip net.IP) error {

	var a socketAddrRequest
	copy(a.name[:], name)
	a.addr = unix.RawSockaddrInet4{}
	a.addr.Family = unix.AF_INET
	copy(a.addr.Addr[:], ip.To4())

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCSIFADDR),
		uintptr(unsafe.Pointer(&a)),
	)
	if errno < 0 {
		log.Println("set ip failed", errno)
		return errors.New("set ip failed")
	}

	return nil
}

func setTunMask(name string, fd int, mask net.IP) error {
	var a socketAddrRequest
	copy(a.name[:], name)
	a.addr = unix.RawSockaddrInet4{}
	a.addr.Family = unix.AF_INET
	copy(a.addr.Addr[:], mask.To4())

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCSIFNETMASK),
		uintptr(unsafe.Pointer(&a)),
	)
	if errno < 0 {
		log.Println("set mask failed", errno)
		return errors.New("set mask failed")
	}

	return nil
}

func configTun(name string, addr, mask net.IP, mtu int) error {

	fd, err := unix.Socket(
		unix.AF_INET,
		unix.SOCK_DGRAM,
		0,
	)

	if err != nil {
		return err
	}

	defer func() {
		unix.Close(fd)
	}()

	// set mtu
	if err = setMTU(name, fd, mtu); err != nil {
		return err
	}

	// set tun ip address
	if err = setTunIP(name, fd, addr); err != nil {
		return err
	}

	// set tun mask
	if err = setTunMask(name, fd, mask); err != nil {
		return err
	}

	// up
	if err = setTunFlags(name, fd, 1, 0); err != nil {
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
