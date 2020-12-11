package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"quicvpn/tun"
	"syscall"

	"github.com/lucas-clemente/quic-go"
)

func Hello(a int, s string) {

	fmt.Println("im in hello!!!", s, a)
}

func Hello2(a int, s string) {

	fmt.Println("im in hello222!!!", s, a)
}

func Hello2tamp(a int, s string) {

	fmt.Println("im in hello!!!", s, a)
}

func SocketTrampoline(domain, typ, proto int) (fd int, err error) {

	return 0, nil
}

var count int

func MySocket(domain, typ, proto int) (fd int, err error) {

	count++
	fmt.Println("current ---------------begin", count)

	if syscall.SOCK_DGRAM&typ != 0 {
		fmt.Println("hook!!!00000000000000000000000")
	}

	fd, err = SocketTrampoline(domain, typ, proto)

	fmt.Println("current is success!!++++++end", fd, err, count)
	fmt.Println("")
	return fd, err
}

const addr = "localhost:4242"

const message = "foobar"

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {

	tdev := tun.NewTun(1500)
	err := tdev.CreateTun("ev0", 1500, net.IPv4(192, 168, 1, 2), net.IPv4(255, 255, 255, 0))
	fmt.Println("err", err)
	if err == nil {

		for {

			buffer, err := tdev.Read()

			fmt.Println(len(buffer), err)
		}

	}

	//go func() { log.Fatal(echoServer()) }()

	//gohook.Hook(syscall.Socket, MySocket, SocketTrampoline)

	//err := clientMain()
	//if err != nil {
	//	panic(err)
	//}

	//Hello2(5, "hello")
	//fmt.Println("--------------------------")
	//Hello(3, "ieu")
	//Hello2tamp(3, "fuck you!!!")

}

// Start a server that echos all data on the first stream opened by the client
func echoServer() error {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		return err
	}
	sess, err := listener.Accept(context.Background())
	if err != nil {
		return err
	}
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		panic(err)
	}
	// Echo through the loggingWriter
	_, err = io.Copy(loggingWriter{stream}, stream)
	return err
}

func clientMain() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
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

	return nil
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
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
		NextProtos:   []string{"quic-echo-example"},
	}
}
