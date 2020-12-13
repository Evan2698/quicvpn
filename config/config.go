package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type VPNSetting struct {
	Server    string `json:"server"`
	Port      uint16 `json:"port"`
	TunServer string `json:"tunserver"`
	TunLocal  string `json:"tunlocal"`
	Pass      string `json:"pass"`
	Mask      string `json:"mask"`
	Tun       string `json:"tun"`
}

func Parse(path string) (config *VPNSetting, err error) {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	config = &VPNSetting{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (s *VPNSetting) Dump() {
	log.Println("server :", s.Server)
	log.Println("server_port :", s.Port)
	log.Println("tunlocal :", s.TunLocal)
	log.Println("tunserver :", s.TunServer)
	log.Println("mask :", s.Mask)
	log.Println("password :", s.Pass)
	log.Println("tun :", s.Tun)
}
