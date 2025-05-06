package rotor

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

var (
	ErrGroupEmpty           = fmt.Errorf("group is empty")
	ErrSubjectEmpty         = fmt.Errorf("subject is empty")
	ErrInvalidMessageFormat = fmt.Errorf("invalid message format")
)

type Message struct {
	Group    Group
	Subject  Subject
	Data     []byte
	Interval time.Duration
	lastSent time.Time
}

func (m *Message) Validate() error {
	if m.Group == "" {
		return ErrGroupEmpty
	}

	if len(m.Subject.Parts) == 0 {
		return ErrSubjectEmpty
	}

	return nil
}

func ParseMessage(payload []byte) (*Message, error) {
	parts := bytes.SplitN(payload, []byte("\\0"), 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid message format")
	}

	group := Group(parts[0])

	subject, err := ParseSubject(string(parts[1]))
	if err != nil {
		return nil, err
	}

	if subject.HasWildcard() {
		return nil, fmt.Errorf("wildcard in subject not allowed")
	}

	data := parts[2]

	return &Message{
		Group:    group,
		Subject:  subject,
		Data:     data,
		Interval: time.Second,
	}, nil
}

func (m *Message) NextTime() time.Time {
	return m.lastSent.Add(m.Interval)
}

func (m *Message) Send(conn *ipv4.PacketConn, addr net.Addr) error {
	p := [][]byte{
		[]byte(m.Group + "\\0"),
		[]byte(m.Subject.String() + "\\0"),
		m.Data,
	}

	payload := bytes.Join(p, []byte{})

	if _, err := conn.WriteTo(payload, nil, addr); err != nil {
		return err
	}

	m.lastSent = time.Now()

	return nil
}
