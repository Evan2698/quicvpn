package server

import (
	"context"
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

func VpnServer(c config.VPNSetting) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()

	tdev, err := common.CreateTun(c.Tun, c.TunServer, c.Mask, common.MTU)
	if err != nil {
		log.Println("create tun device failed,", err)
		return err
	}

	defer tdev.Close()

	svAddress := net.JoinHostPort(c.Server, strconv.Itoa(int(c.Port)))

	listener, err := quic.ListenAddr(svAddress, common.GenerateTLSConfig(c.Pass), nil)
	if err != nil {
		return err
	}

	defer listener.Close()

	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			log.Println("session listen failed!!!", err)
			break
		}
		go handleSession(sess, tdev)
	}

	return nil

}

func handleTun2Remote(s quic.Stream, dev tun.Tun, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()
	defer wg.Done()

	for {
		buffer, err := dev.Read()
		if err != nil {
			log.Println("read tun is failed!!!")
			break
		}

		ulen := len(buffer)
		vLenBuffer := utils.Int2Bytes(uint32(ulen))
		_, err = common.WriteXBytes(vLenBuffer, s)
		if err != nil {
			log.Println("write length failed!!")
			break
		}
		_, err = common.WriteXBytes(buffer, s)
		if err != nil {
			log.Println("write content to remote failed!!")
			break
		}
	}

}

func handleRemote2Tun(s quic.Stream, dev tun.Tun, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()
	defer wg.Done()

	buffer := make([]byte, common.MTU)
	defer func() { buffer = nil }()

	for {

		content, err := common.ReadXBytes(4, buffer[:4], s)
		if err != nil {
			log.Println("read raw content failed", err)
			break
		}

		vLen := utils.Bytes2Int(content)
		log.Println("read: ", content, vLen)

		if vLen > common.MTU {
			log.Println("content length is too long", vLen)
			break
		}

		buffer, err = common.ReadXBytes(vLen, buffer, s)
		if err != nil {
			log.Println("read content failed: ", err)
			break
		}

		dev.Write(buffer[:vLen])
	}

}

func handleSession(sess quic.Session, dev tun.Tun) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()

	defer sess.CloseWithError(0x34, "Error ocurred!!!")
	for {
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			stream.Close()
			log.Println("Accept Stream failed!!!", err)
			break
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go handleRemote2Tun(stream, dev, &wg)
		go handleTun2Remote(stream, dev, &wg)

		wg.Wait()
		stream.Close()
	}

}
