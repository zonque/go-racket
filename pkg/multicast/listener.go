package multicast

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"sync"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/ipv4"
)

const (
	maxMTU = 1500
)

type listener struct {
	mutex sync.Mutex

	ipv4PacketConn *ipv4.PacketConn
	ifis           []*net.Interface

	streams map[string]consumers
}

func (l *listener) close() {
	if err := l.ipv4PacketConn.Close(); err != nil {
		log.Warn().Err(err).Msg("failed to close ipv4 packet conn")
	}
}

func (l *listener) addConsumer(c *Consumer) error {
	k := c.addr.IP.String()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, ok := l.streams[k]; !ok {
		l.streams[k] = make(consumers, 0)

		for _, ifi := range l.ifis {
			if err := l.ipv4PacketConn.JoinGroup(ifi, c.addr); err != nil {
				return fmt.Errorf("failed to join group %s on %s: %w", c.addr, ifi.Name, err)
			}
		}
	}

	l.streams[k] = append(l.streams[k], c)

	log.Info().Str("addr", c.addr.String()).Msg("added consumer")

	return nil
}

func (l *listener) removeConsumer(c *Consumer) {
	k := c.addr.String()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	cs, ok := l.streams[k]
	if !ok {
		return
	}

	for i, cc := range cs {
		if cc == c {
			l.streams[k] = slices.Delete(cs, i, i+1)

			break
		}
	}

	if len(l.streams[k]) == 0 {
		delete(l.streams, k)

		for _, ifi := range l.ifis {
			if err := l.ipv4PacketConn.LeaveGroup(ifi, c.addr); err != nil {
				log.Error().Err(err).Msg("failed to leave group")
			}
		}
	}
}

func (l *listener) hasConsumers() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.streams) > 0
}

func newListener(port int, ifis []*net.Interface) (*listener, error) {
	pc, err := OpenPacketConn(port, "")
	if err != nil {
		return nil, fmt.Errorf("failed to open packet conn: %w", err)
	}

	if err := pc.SetControlMessage(ipv4.FlagDst, true); err != nil {
		return nil, fmt.Errorf("failed to set control message: %w", err)
	}

	l := &listener{
		ipv4PacketConn: pc,
		streams:        make(map[string]consumers),
		ifis:           ifis,
	}

	go func() {
		buf := make([]byte, maxMTU)

		for {
			n, cm, _, err := pc.ReadFrom(buf)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Error().Err(err).Msg("failed to read from packet conn")
				}

				return
			}

			k := cm.Dst.String()

			l.mutex.Lock()

			if cs, ok := l.streams[k]; ok {
				for _, c := range cs {
					newBuf := make([]byte, n)
					copy(newBuf, buf[:n])

					go c.cb(newBuf)
				}
			}

			l.mutex.Unlock()
		}
	}()

	return l, nil
}
