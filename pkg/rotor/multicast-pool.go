package rotor

import (
	"crypto/sha256"
	"net"
	"strings"

	"golang.org/x/net/ipv4"
)

const (
	port = 19090
)

type MulticastPool struct {
	base      net.IPNet
	receivers map[string]*ipv4.PacketConn
	// senders   map[string]*net.UDPConn
	senders map[string]*ipv4.PacketConn
}

func NewMulticastPool(base net.IPNet) *MulticastPool {
	return &MulticastPool{
		base:      base,
		receivers: make(map[string]*ipv4.PacketConn),
		// senders:   make(map[string]*net.UDPConn),
		senders: make(map[string]*ipv4.PacketConn),
	}
}

func (m *MulticastPool) AddressForSubject(subject Subject) *net.UDPAddr {
	s := strings.Join(subject.Parts[:subject.GroupDepth], ".")
	h := sha256.Sum256([]byte(s))

	a := [4]byte{m.base.IP[12], m.base.IP[13], m.base.IP[14], m.base.IP[15]}

	for i, mb := range m.base.Mask {
		a[i] |= h[i] & ^mb
	}

	udpAddr := &net.UDPAddr{
		IP:   net.IP(a[:]),
		Port: port,
	}

	return udpAddr
}
