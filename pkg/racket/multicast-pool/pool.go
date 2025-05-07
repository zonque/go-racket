package multicastpool

import (
	"crypto/sha256"
	"net"

	"github.com/holoplot/go-racket/pkg/racket/global"
	"github.com/holoplot/go-racket/pkg/racket/stream"
)

type Pool struct {
	base net.IPNet
}

func New(base net.IPNet) *Pool {
	return &Pool{
		base: base,
	}
}

func (m *Pool) AddressForStream(stream stream.Stream) *net.UDPAddr {
	h := sha256.Sum256([]byte(stream))

	ip := make([]byte, 4)
	copy(ip, m.base.IP.To4())

	for i, mb := range m.base.Mask {
		ip[i] |= h[i] & ^mb
	}

	udpAddr := &net.UDPAddr{
		IP:   ip,
		Port: global.DefaultPort,
	}

	return udpAddr
}
