package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"time"

	"github.com/holoplot/go-rotor/pkg/rotor"
	"github.com/holoplot/go-rotor/pkg/rotor/group"
	"github.com/holoplot/go-rotor/pkg/rotor/message"
	"github.com/holoplot/go-rotor/pkg/rotor/subject"
)

func randomBytes(size int) []byte {
	b := make([]byte, size)

	for i := 0; i < size; i++ {
		b[i] = byte(rand.IntN(256))
	}

	return b
}

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

	eth, err := net.InterfaceByName("enp0s31f6")
	if err != nil {
		panic(err)
	}

	ifis := []*net.Interface{lo, eth}

	sender := rotor.NewSender(ifis, multicastPool)

	n := 0

	for groupIndex := range 256 {
		g := group.Group(fmt.Sprintf("group-%d", groupIndex))

		for subjectIndex := range 1024 {
			s := subject.Subject{
				Parts: []string{"org", "holoplot", fmt.Sprintf("rotor-%d", subjectIndex)},
			}

			msg := &message.Message{
				Group:    g,
				Subject:  s,
				Data:     randomBytes(128),
				Interval: time.Millisecond*time.Duration(rand.IntN(1000)) + time.Second,
			}

			if err := sender.Publish(msg); err != nil {
				panic(err)
			}

			n++
		}
	}

	log.Printf("Published %d messages\n", n)

	time.Sleep(time.Minute)
	sender.Flush()
}
