package message

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/holoplot/go-racket/pkg/racket/stream"
	"github.com/holoplot/go-racket/pkg/racket/subject"
	"golang.org/x/net/ipv4"
)

var (
	ErrStreamEmpty          = fmt.Errorf("stream is empty")
	ErrSubjectEmpty         = fmt.Errorf("subject is empty")
	ErrInvalidMessageFormat = fmt.Errorf("invalid message format")
	ErrInvalidMessageSize   = fmt.Errorf("invalid message size")
)

type Message struct {
	mutex sync.Mutex

	Stream    stream.Stream
	Subject   subject.Subject
	Data      []byte
	Interval  time.Duration
	hash      string
	timestamp []byte
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

func makeTimestamp() []byte {
	t := time.Now().UnixMicro()
	b := bytes.NewBuffer(nil)

	if err := binary.Write(b, binary.BigEndian, t); err != nil {
		panic(err)
	}

	return b.Bytes()
}

func Parse(payload []byte) (*Message, error) {
	if len(payload) < 8 {
		return nil, ErrInvalidMessageSize
	}

	timestamp := payload[:8]
	payload = payload[8:]

	parts := bytes.SplitN(payload, []byte("\\0"), 3)
	if len(parts) != 3 {
		return nil, ErrInvalidMessageFormat
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
		Stream:    stream,
		Subject:   subject,
		Data:      data,
		Interval:  time.Second,
		timestamp: timestamp,
	}, nil
}

func (m *Message) Hash() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.hash != "" {
		return m.hash
	}

	h := sha256.New()
	h.Write([]byte(m.Stream))
	h.Write([]byte(m.Subject.String()))
	h.Write(m.Data)

	m.hash = hex.EncodeToString(h.Sum(nil))

	return m.hash
}

func (m *Message) TimeStamp() time.Time {
	m.mutex.Lock()
	if len(m.timestamp) == 0 {
		m.timestamp = makeTimestamp()
	}
	m.mutex.Unlock()

	var t int64
	if err := binary.Read(bytes.NewReader(m.timestamp), binary.BigEndian, &t); err != nil {
		panic(err)
	}

	return time.UnixMicro(t)
}

func (m *Message) Send(conn *ipv4.PacketConn, addr net.Addr) error {
	m.mutex.Lock()
	if len(m.timestamp) == 0 {
		m.timestamp = makeTimestamp()
	}
	m.mutex.Unlock()

	p := [][]byte{
		[]byte(m.timestamp),
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
