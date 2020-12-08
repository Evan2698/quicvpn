package main

import (
	"log"
	"net"
	"quicvpn/logex"
	"quicvpn/tun"
	"sync"
)

func main() {
	logex.LOGHIDE = 0
	log.Print("xxxxxx")

	tun := tun.NewTun(1500)

	err := tun.OpenTun(net.IPv4(10, 0, 0, 1), net.IPv4(10, 0, 0, 0), net.IPv4(255, 255, 255, 0))
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	defer tun.Close()

	wg.Add(1)

	go func() {

		for {
			by, err := tun.Read()
			if err != nil {
				log.Print("looooo", err)
				panic(err)
			}

			log.Print("read bytes: ", len(by))
			log.Println("src:", net.IP(by[12:16]), "dst:", net.IP(by[16:20]))
		}

		wg.Done()

	}()

	wg.Wait()

}
