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

	s2 := rotor.Subject{
		Parts:      []string{"org", "holoplot", "go", "rotor", "blah"},
		GroupDepth: 3,
	}

	a := multicastPool.AddressForSubject(s)
	println(a.String())

	sender := rotor.NewSender(multicastPool)

	go sender.Run()

	msg := &rotor.Message{
		Subject:  s,
		Data:     []byte("Hello, world!"),
		Interval: time.Second,
	}

	msg2 := &rotor.Message{
		Subject:  s2,
		Data:     []byte("fck ths sht!"),
		Interval: time.Second / 2,
	}

	if err := sender.Publish(msg); err != nil {
		panic(err)
	}

	if err := sender.Publish(msg2); err != nil {
		panic(err)
	}

	time.Sleep(10 * time.Second)
	sender.Flush()
}
