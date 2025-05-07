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

	ip := make([]byte, 4)
	copy(ip, m.base.IP.To4())

	for i, mb := range m.base.Mask {
		ip[i] |= h[i] & ^mb
	}

	udpAddr := &net.UDPAddr{
		IP:   ip,
		Port: defaultPort,
	}

	return udpAddr
}
