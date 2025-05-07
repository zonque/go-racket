package racket

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/holoplot/go-racket/pkg/multicast"
	"github.com/holoplot/go-racket/pkg/racket/message"
	"github.com/holoplot/go-racket/pkg/racket/stream"
	"golang.org/x/net/ipv4"
)

type Sender struct {
	lock sync.RWMutex

	ifis          []*net.Interface
	pool          *MulticastPool
	senderStreams map[stream.Stream]*senderStream
}

type queuedMessage struct {
	msg    *message.Message
	cancel context.CancelFunc
}

type senderStream struct {
	lock     sync.RWMutex
	sendLock sync.Mutex
	pool     *MulticastPool
	pcs      []*ipv4.PacketConn
	messages map[string]*queuedMessage
}

func newSenderStream(ifis []*net.Interface, pool *MulticastPool) (*senderStream, error) {
	sg := &senderStream{
		pool:     pool,
		pcs:      make([]*ipv4.PacketConn, 0),
		messages: make(map[string]*queuedMessage),
	}

	if err := sg.reopenPacketConns(ifis); err != nil {
		return nil, err
	}

	return sg, nil
}

func (sg *senderStream) reopenPacketConns(ifis []*net.Interface) error {
	sg.lock.Lock()
	defer sg.lock.Unlock()

	for _, pc := range sg.pcs {
		if err := pc.Close(); err != nil {
			return err
		}
	}

	pcs, err := multicast.OpenPacketConns(ifis, defaultPort)
	if err != nil {
		return err
	}

	sg.pcs = pcs

	return nil
}

func (sg *senderStream) send(m *message.Message, addr *net.UDPAddr) error {
	sg.sendLock.Lock()
	defer sg.sendLock.Unlock()

	for _, pc := range sg.pcs {
		if err := m.Send(pc, addr); err != nil {
			return err
		}
	}

	return nil
}

func (sg *senderStream) publish(m *message.Message) {
	ctx, cancel := context.WithCancel(context.TODO())

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

		addr := sg.pool.AddressForStream(m.Stream)

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
					if ctx.Err() != nil {
						return
					}

					fmt.Printf("Error sending message: %v\n", err)
				}
			}
		}
	}()
}

func (sg *senderStream) flush() {
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

func NewSender(ifis []*net.Interface, pool *MulticastPool) (*Sender, error) {
	sender := &Sender{
		senderStreams: make(map[stream.Stream]*senderStream),
		ifis:          ifis,
		pool:          pool,
	}

	return sender, nil
}

func (s *Sender) Publish(m *message.Message) error {
	if err := m.Validate(); err != nil {
		return err
	}

	if m.Subject.HasWildcard() {
		return fmt.Errorf("wildcard in subject not allowed")
	}

	s.lock.Lock()

	sg := s.senderStreams[m.Stream]
	if sg == nil {
		var err error

		sg, err = newSenderStream(s.ifis, s.pool)
		if err != nil {
			s.lock.Unlock()
			return err
		}

		s.senderStreams[m.Stream] = sg
	}

	s.lock.Unlock()

	sg.publish(m)

	return nil
}

func (s *Sender) Flush() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, sg := range s.senderStreams {
		sg.flush()
	}

	s.senderStreams = make(map[stream.Stream]*senderStream)
}
