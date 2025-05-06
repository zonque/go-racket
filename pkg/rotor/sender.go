package rotor

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/holoplot/go-rotor/pkg/multicast"
	"github.com/holoplot/go-rotor/pkg/rotor/group"
	"github.com/holoplot/go-rotor/pkg/rotor/message"
	"golang.org/x/net/ipv4"
)

type Sender struct {
	lock sync.RWMutex

	ifis         []*net.Interface
	pool         *MulticastPool
	senderGroups map[group.Group]*senderGroup
}

type queuedMessage struct {
	msg    *message.Message
	cancel context.CancelFunc
}

type senderGroup struct {
	lock     sync.RWMutex
	sendLock sync.Mutex
	pool     *MulticastPool
	pcs      []*ipv4.PacketConn
	messages map[string]*queuedMessage
}

func newSenderGroup(ifis []*net.Interface, pool *MulticastPool) (*senderGroup, error) {
	pcs, err := multicast.OpenPacketConns(ifis, defaultPort)
	if err != nil {
		return nil, err
	}

	return &senderGroup{
		pcs:      pcs,
		pool:     pool,
		messages: make(map[string]*queuedMessage),
	}, nil
}

func (sg *senderGroup) send(m *message.Message, addr *net.UDPAddr) error {
	sg.sendLock.Lock()
	defer sg.sendLock.Unlock()

	for _, pc := range sg.pcs {
		if err := m.Send(pc, addr); err != nil {
			return err
		}
	}

	return nil
}

func (sg *senderGroup) publish(m *message.Message) {
	ctx, cancel := context.WithCancel(context.Background())

	sg.lock.Lock()

	if qm, ok := sg.messages[m.Subject.String()]; ok {
		qm.cancel()
	}

	sg.messages[m.Subject.String()] = &queuedMessage{
		msg:    m,
		cancel: cancel,
	}

	sg.lock.Unlock()

	go func() {
		ticker := time.NewTicker(m.Interval)
		defer ticker.Stop()

		addr := sg.pool.AddressForGroup(m.Group)

		// Send the message immediately
		if err := sg.send(m, addr); err != nil {
			fmt.Printf("Error sending message: %v\n", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := sg.send(m, addr); err != nil {
					fmt.Printf("Error sending message: %v\n", err)
				}
			}
		}
	}()
}

func (sg *senderGroup) flush() {
	sg.lock.Lock()
	defer sg.lock.Unlock()

	for _, qm := range sg.messages {
		qm.cancel()
	}

	for _, pc := range sg.pcs {
		pc.Close()
	}

	sg.messages = make(map[string]*queuedMessage)
}

func NewSender(ifis []*net.Interface, pool *MulticastPool) *Sender {
	return &Sender{
		senderGroups: make(map[group.Group]*senderGroup),
		ifis:         ifis,
		pool:         pool,
	}
}

func (s *Sender) Publish(m *message.Message) error {
	if err := m.Validate(); err != nil {
		return err
	}

	if m.Subject.HasWildcard() {
		return fmt.Errorf("wildcard in subject not allowed")
	}

	s.lock.Lock()
	sg := s.senderGroups[m.Group]
	if sg == nil {
		var err error

		sg, err = newSenderGroup(s.ifis, s.pool)
		if err != nil {
			s.lock.Unlock()
			return err
		}

		s.senderGroups[m.Group] = sg
	}

	sg.publish(m)

	defer s.lock.Unlock()

	return nil
}

func (s *Sender) Flush() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, sg := range s.senderGroups {
		sg.flush()
	}

	s.senderGroups = make(map[group.Group]*senderGroup)
}
