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

	groups map[string]consumers
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

	if _, ok := l.groups[k]; !ok {
		l.groups[k] = make(consumers, 0)

		for _, ifi := range l.ifis {
			if err := l.ipv4PacketConn.JoinGroup(ifi, c.addr); err != nil {
				return fmt.Errorf("failed to join group %s on %s: %w", c.addr, ifi.Name, err)
			}
		}
	}

	l.groups[k] = append(l.groups[k], c)

	log.Info().Str("addr", c.addr.String()).Msg("added consumer")

	return nil
}

func (l *listener) removeConsumer(c *Consumer) {
	k := c.addr.String()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	cs, ok := l.groups[k]
	if !ok {
		return
	}

	for i, cc := range cs {
		if cc == c {
			l.groups[k] = slices.Delete(cs, i, i+1)

			break
		}
	}

	if len(l.groups[k]) == 0 {
		delete(l.groups, k)

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

	return len(l.groups) > 0
}

func newListener(port int, ifis []*net.Interface) (*listener, error) {
	pc, err := OpenPacketConn(net.IPv4(0, 0, 0, 0), port, "")
	if err != nil {
		return nil, fmt.Errorf("failed to open packet conn: %w", err)
	}

	if err := pc.SetControlMessage(ipv4.FlagDst, true); err != nil {
		return nil, fmt.Errorf("failed to set control message: %w", err)
	}

	l := &listener{
		ipv4PacketConn: pc,
		groups:         make(map[string]consumers),
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

			if cs, ok := l.groups[k]; ok {
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
