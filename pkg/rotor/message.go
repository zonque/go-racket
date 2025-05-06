package rotor

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

type Message struct {
	Subject  Subject
	Data     []byte
	Interval time.Duration
	lastSent time.Time
}

func ParseMessage(payload []byte) (*Message, error) {
	parts := bytes.SplitN(payload, []byte("\\0"), 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid message format")
	}

	subject := ParseSubject(string(parts[0]))
	data := parts[1]

	return &Message{
		Subject:  subject,
		Data:     data,
		Interval: time.Second,
	}, nil
}

func (m *Message) NextTime() time.Time {
	return m.lastSent.Add(m.Interval)
}

func (m *Message) Send(conn *ipv4.PacketConn, addr net.Addr) error {
	payload := append([]byte(m.Subject.String()+"\\0"), m.Data...)

	if _, err := conn.WriteTo(payload, nil, addr); err != nil {
		return err
	}

	m.lastSent = time.Now()

	return nil
}
