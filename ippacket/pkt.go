package ippacket

import (
	"errors"
	"net"
)

type Packet interface {
	IsV4() bool
	IsV6() bool
	Src() net.IP
	Dst() net.IP
	Which() uint16
}

type packetContent struct {
	SrcAddress net.IP
	DstAddress net.IP
	I4         bool
	Protocol   uint16
}

func TryParse(stream []byte) (Packet, error) {

	p := &packetContent{
		I4: true,
	}

	if len(stream) < 20 {
		return nil, errors.New("not ipv4")
	}

	if (stream[0] & 0xf0) == 0x60 {
		p.I4 = false
		if len(stream) < 40 {
			return nil, errors.New("not ipv6")
		}
		p.SrcAddress = net.IP(stream[8:24])
		p.DstAddress = net.IP(stream[24:40])
		p.Protocol = uint16(stream[6])

	} else {
		p.SrcAddress = net.IP(stream[12:16])
		p.DstAddress = net.IP(stream[16:20])
		p.Protocol = uint16(stream[9])
	}

	return p, nil
}

func (p *packetContent) IsV4() bool {

	return p.I4
}

func (p *packetContent) IsV6() bool {

	return !p.I4
}

func (p *packetContent) Src() net.IP {
	return p.SrcAddress
}
func (p *packetContent) Dst() net.IP {
	return p.DstAddress
}
func (p *packetContent) Which() uint16 {
	return p.Protocol
}
