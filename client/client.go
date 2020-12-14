package client

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"quicvpn/common"
	"quicvpn/config"
	"quicvpn/tun"
	"quicvpn/utils"
	"strconv"
	"sync"

	"github.com/lucas-clemente/quic-go"
)

func LauchClient(fd int, setting config.VPNSetting) {

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{setting.Pass},
	}

	addr := net.JoinHostPort(setting.Server, strconv.Itoa(int(setting.Port)))
	dev, _ := tun.CreateAndroidTunDevice(fd)

	var w sync.WaitGroup
	var e, e1 error

	for {
		session, err := quic.DialAddr(addr, tlsConf, nil)
		if err != nil && e == nil && e1 == nil {
			log.Println("create session failed", err)
			continue
		}
		stream, err := session.OpenStreamSync(context.Background())
		if err != nil && e == nil && e1 == nil {
			log.Println("open stream failed", err)
			continue
		}

		w.Add(1)
		go func() {
			e = recvfromRemote(stream, dev)
			w.Done()

		}()

		e1 = send2remote(stream, dev)
		if e1.Error() != "1" && e.Error() != "1" {
			stream.Close()
		}

		w.Wait()

		if e1.Error() == "1" || e.Error() == "1" {
			log.Println("There is a fetal error on tun device!!")
			break
		}
		e1 = nil
		e = nil

		session.CloseWithError(23, "xxxx")
	}

}

func recvfromRemote(s quic.Stream, dev tun.Tun) error {

	buffer := make([]byte, tun.MTU)
	for {

		content, err := common.ReadXBytes(4, buffer[:4], s)
		if err != nil {
			log.Println("read raw content failed", err)
			return errors.New("3")
		}

		vLen := utils.Bytes2Int(content)
		log.Println("read: ", content, vLen)

		if vLen > tun.MTU {
			log.Println("content length is too long", vLen)
			return errors.New("2")
		}

		buffer, err = common.ReadXBytes(vLen, buffer, s)
		if err != nil {
			log.Println("read content failed: ", err)
			return errors.New("4")
		}

		_, err = dev.Write(buffer[:vLen])
		if err != nil {
			return errors.New("1")
		}
	}
}

func send2remote(s quic.Stream, dev tun.Tun) error {

	for {

		content, err := dev.Read()
		if err != nil {
			log.Println("read from tun failed!!!")
			return errors.New("1")
		}

		ulen := len(content)
		vLenBuffer := utils.Int2Bytes(uint32(ulen))
		_, err = common.WriteXBytes(vLenBuffer, s)
		if err != nil {
			log.Println("write length failed!!")
			return errors.New("2")
		}
		_, err = common.WriteXBytes(content, s)
		if err != nil {
			log.Println("write content to remote failed!!")
			return errors.New("3")
		}

	}
}
