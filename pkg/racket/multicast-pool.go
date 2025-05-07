package racket

import (
	"crypto/sha256"
	"net"

	"github.com/holoplot/go-racket/pkg/racket/stream"
)

const (
	defaultPort = 19090
)

type MulticastPool struct {
	base net.IPNet
}

func NewMulticastPool(base net.IPNet) *MulticastPool {
	return &MulticastPool{
		base: base,
	}
}

func (m *MulticastPool) AddressForStream(stream stream.Stream) *net.UDPAddr {
	h := sha256.Sum256([]byte(stream))

	a := [4]byte{m.base.IP[12], m.base.IP[13], m.base.IP[14], m.base.IP[15]}

	for i, mb := range m.base.Mask {
		a[i] |= h[i] & ^mb
	}

	udpAddr := &net.UDPAddr{
		IP:   net.IP(a[:]),
		Port: defaultPort,
	}

	return udpAddr
}
