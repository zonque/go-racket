package main

import (
	"net"
	"time"

	"github.com/holoplot/go-racket/pkg/racket"
	"github.com/holoplot/go-racket/pkg/racket/message"
	"github.com/holoplot/go-racket/pkg/racket/subject"
)

func main() {
	base := net.IPNet{
		IP:   net.ParseIP("239.0.0.0"),
		Mask: net.CIDRMask(16, 32),
	}

	multicastPool := racket.NewMulticastPool(base)

	lo, err := net.InterfaceByName("lo")
	if err != nil {
		panic(err)
	}

	ifis := []*net.Interface{lo}

	s := subject.Subject{
		Parts: []string{"org", "holoplot", "go", "racket", "demo"},
	}

	sender, err := racket.NewSender(ifis, multicastPool)
	if err != nil {
		panic(err)
	}

	msg := &message.Message{
		Subject:  s,
		Data:     []byte("Hello, world!"),
		Interval: time.Second,
	}

	if err := sender.Publish(msg); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)
	sender.Flush()
}
