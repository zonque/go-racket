package rotor

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/holoplot/go-rotor/pkg/multicast"
	"golang.org/x/net/ipv4"
)

type Sender struct {
	lock      sync.RWMutex
	messages  map[string]*Message
	triggerCh chan struct{}
	pool      *MulticastPool
	pc        *ipv4.PacketConn
}

func NewSender(pool *MulticastPool) *Sender {
	pc, err := multicast.OpenPacketConn(net.IP{127, 0, 0, 1}, 19090, "")
	if err != nil {
		panic(err)
	}

	return &Sender{
		messages:  make(map[string]*Message),
		triggerCh: make(chan struct{}),
		pool:      pool,
		pc:        pc,
	}
}

func (s *Sender) nextMessage() *Message {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// Iterate over all the messages and find the one that should be sent
	// next based on the interval and last sent time.
	var next *Message

	for _, m := range s.messages {
		if next == nil {
			next = m
			continue
		}

		if m.NextTime().Before(next.NextTime()) {
			next = m
		}
	}

	return next
}

func (s *Sender) SendMessage(m *Message) error {
	if err := m.Validate(); err != nil {
		return err
	}

	if m.Subject.HasWildcard() {
		return fmt.Errorf("wildcard in subject not allowed")
	}

	addr := s.pool.AddressForGroup(m.Group)

	log.Printf("Sending message to %v: %s", addr, m.Data)

	if err := m.Send(s.pc, addr); err != nil {
		return err
	}

	return nil
}

func (s *Sender) Run() {
	for {
		next := s.nextMessage()
		if next == nil {
			<-s.triggerCh

			continue
		}

		log.Printf("Next message to send at %v, now is %v", next.lastSent, time.Now())

		select {
		case <-s.triggerCh:
		case <-time.After(time.Until(next.NextTime())):
			if err := s.SendMessage(next); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}

func (s *Sender) Publish(m *Message) error {
	s.lock.Lock()
	s.messages[m.Subject.String()] = m
	s.lock.Unlock()

	s.triggerCh <- struct{}{}

	return nil
}

func (s *Sender) Flush() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.messages = make(map[string]*Message)
}
