package main

import (
	"net"
	"time"

	"github.com/holoplot/go-rotor/pkg/rotor"
	"github.com/holoplot/go-rotor/pkg/rotor/message"
	"github.com/holoplot/go-rotor/pkg/rotor/subject"
)

func main() {
	base := net.IPNet{
		IP:   net.ParseIP("239.0.0.0"),
		Mask: net.CIDRMask(16, 32),
	}

	multicastPool := rotor.NewMulticastPool(base)

	lo, err := net.InterfaceByName("lo")
	if err != nil {
		panic(err)
	}

	ifis := []*net.Interface{lo}

	s := subject.Subject{
		Parts: []string{"org", "holoplot", "go", "rotor", "demo"},
	}

	sender := rotor.NewSender(ifis, multicastPool)

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
