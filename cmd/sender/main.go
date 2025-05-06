package main

import (
	"net"
	"time"

	"github.com/holoplot/go-rotor/pkg/rotor"
)

func main() {
	base := net.IPNet{
		IP:   net.ParseIP("239.0.0.0"),
		Mask: net.CIDRMask(16, 32),
	}

	multicastPool := rotor.NewMulticastPool(base)

	s := rotor.Subject{
		Parts:      []string{"org", "holoplot", "go", "rotor", "demo"},
		GroupDepth: 3,
	}

	sender := rotor.NewSender(multicastPool)

	go sender.Run()

	msg := &rotor.Message{
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
