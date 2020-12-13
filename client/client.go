package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"quicvpn/config"
	"quicvpn/tun"
	"strconv"

	"github.com/lucas-clemente/quic-go"
)

func LauchClient(fd int, setting config.VPNSetting) {

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{setting.Pass},
	}

	addr := net.JoinHostPort(setting.Server, strconv.Itoa(int(setting.Port)))
	dev, _ := tun.CreateAndroidTunDevice(fd)
	dev.

	for {

	}

	session, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		return err
	}

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("Client: Sending '%s'\n", message)
	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}

	buf := make([]byte, len(message))
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		return err
	}
	fmt.Printf("Client: Got '%s'\n", buf)

}
