package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"log"
	"math/big"
	"net"
	"quicvpn/tun"
	"quicvpn/utils"
)

func CreateTun(name, ip, mask string, mtu int) (tun.Tun, error) {

	addr := net.ParseIP(ip)
	ma := net.ParseIP(mask)

	tdev, err := tun.CreateTun(name, mtu, addr, ma)
	if err != nil {
		return nil, err
	}

	return tdev, nil
}

func GenerateTLSConfig(pass string) *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{pass},
	}
}

func ReadXBytes(bytes uint32, buffer []byte, con io.ReadWriteCloser) ([]byte, error) {
	defer utils.Trace("readXBytes.readXBytes")()
	if bytes <= 0 {
		return nil, errors.New("0 bytes can not read! ")
	}

	var index uint32
	var err error
	var n int
	for {
		n, err = con.Read(buffer[index:])
		log.Println("read from socket size: ", n, err)
		if err != nil {
			log.Println("error on read_bytes_from_socket ", n, err)
			break
		}
		index = index + uint32(n)

		if index >= bytes {
			log.Println("read count for output ", index, err)
			break
		}
	}
	if index == bytes {
		err = nil
	}

	log.Println("read result size: ", index, err)
	return buffer[:bytes], err
}

func WriteXBytes(buffer []byte, con io.ReadWriteCloser) (int, error) {
	defer utils.Trace("writeXBytes.writeXBytes")()
	nbytes := uint32(len(buffer))
	var index uint32 = 0
	var err error
	var n int
	for {
		n, err = con.Write(buffer[index:])
		if err != nil {
			log.Println("write bytes error! ", n, err)
			break
		}
		index = index + uint32(n)
		if index >= nbytes {
			break
		}
	}
	if index == nbytes {
		err = nil
	}

	log.Println("writeXBytes >>>>>>", n, err)

	return int(index), err
}
