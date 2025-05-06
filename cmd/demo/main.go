package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"time"

	"github.com/holoplot/go-rotor/pkg/rotor"
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

	sender := rotor.NewSender(multicastPool)

	n := 0

	for groupIndex := range 256 {
		for subjectIndex := range 1024 {
			g := rotor.Group(fmt.Sprintf("group-%d", groupIndex))
			s := rotor.Subject{
				Parts: []string{"org", "holoplot", fmt.Sprintf("rotor-%d", subjectIndex)},
			}

			msg := &rotor.Message{
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
