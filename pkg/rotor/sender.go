package rotor

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/holoplot/go-rotor/pkg/multicast"
	"golang.org/x/net/ipv4"
)

type Sender struct {
	lock         sync.RWMutex
	pool         *MulticastPool
	senderGroups map[Group]*senderGroup
}

type queuedMessage struct {
	msg    *Message
	cancel context.CancelFunc
}

type senderGroup struct {
	lock     sync.RWMutex
	sendLock sync.Mutex
	pool     *MulticastPool
	pc       *ipv4.PacketConn
	messages map[string]*queuedMessage
}

func newSenderGroup(pool *MulticastPool) (*senderGroup, error) {
	pc, err := multicast.OpenPacketConn(net.IP{127, 0, 0, 1}, 19090, "")
	if err != nil {
		return nil, err
	}

	return &senderGroup{
		pc:       pc,
		pool:     pool,
		messages: make(map[string]*queuedMessage),
	}, nil
}

func (sg *senderGroup) send(m *Message, addr *net.UDPAddr) error {
	sg.sendLock.Lock()
	defer sg.sendLock.Unlock()

	if err := m.Send(sg.pc, addr); err != nil {
		return err
	}

	return nil
}

func (sg *senderGroup) publish(m *Message) {
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

func NewSender(pool *MulticastPool) *Sender {
	return &Sender{
		senderGroups: make(map[Group]*senderGroup),
		pool:         pool,
	}
}

func (s *Sender) Publish(m *Message) error {
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

		sg, err = newSenderGroup(s.pool)
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

	// for _, qm := range s.messages {
	// 	qm.cancel()
	// }

	// s.messages = make(map[string]*queuedMessage)
}
