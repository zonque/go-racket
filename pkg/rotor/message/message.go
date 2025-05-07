package message

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/holoplot/go-rotor/pkg/rotor/stream"
	"github.com/holoplot/go-rotor/pkg/rotor/subject"
	"golang.org/x/net/ipv4"
)

var (
	ErrStreamEmpty          = fmt.Errorf("stream is empty")
	ErrSubjectEmpty         = fmt.Errorf("subject is empty")
	ErrInvalidMessageFormat = fmt.Errorf("invalid message format")
)

type Message struct {
	Stream   stream.Stream
	Subject  subject.Subject
	Data     []byte
	Interval time.Duration
}

func (m *Message) Validate() error {
	if m.Stream == "" {
		return ErrStreamEmpty
	}

	if len(m.Subject.Parts) == 0 {
		return ErrSubjectEmpty
	}

	return nil
}

func Parse(payload []byte) (*Message, error) {
	parts := bytes.SplitN(payload, []byte("\\0"), 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid message format")
	}

	stream := stream.Stream(parts[0])

	subject, err := subject.Parse(string(parts[1]))
	if err != nil {
		return nil, err
	}

	if subject.HasWildcard() {
		return nil, fmt.Errorf("wildcard in subject not allowed")
	}

	data := parts[2]

	return &Message{
		Stream:   stream,
		Subject:  subject,
		Data:     data,
		Interval: time.Second,
	}, nil
}

func (m *Message) Send(conn *ipv4.PacketConn, addr net.Addr) error {
	p := [][]byte{
		[]byte(m.Stream + "\\0"),
		[]byte(m.Subject.String() + "\\0"),
		m.Data,
	}

	payload := bytes.Join(p, []byte{})

	if _, err := conn.WriteTo(payload, nil, addr); err != nil {
		return err
	}

	return nil
}

func (m *Message) PeriodicSend(ctx context.Context, conn *ipv4.PacketConn, addr net.Addr) {
	ticker := time.NewTicker(m.Interval)
	defer ticker.Stop()

	// Send the message immediately
	if err := m.Send(conn, addr); err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.Send(conn, addr); err != nil {
				fmt.Printf("Error sending message: %v\n", err)
			}
		}
	}
}
