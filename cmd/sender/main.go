package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/holoplot/go-racket/pkg/racket/message"
	multicastpool "github.com/holoplot/go-racket/pkg/racket/multicast-pool"
	racket "github.com/holoplot/go-racket/pkg/racket/sender"
	"github.com/holoplot/go-racket/pkg/racket/stream"
	"github.com/holoplot/go-racket/pkg/racket/subject"
)

func randomBytes(size int) []byte {
	b := make([]byte, size)

	for i := 0; i < size; i++ {
		b[i] = byte(rand.IntN(256))
	}

	return b
}

func main() {
	_, base, _ := net.ParseCIDR("239.0.0.0/16")
	multicastPool, err := multicastpool.New(*base)
	if err != nil {
		panic(err)
	}

	lo, err := net.InterfaceByName("lo")
	if err != nil {
		panic(err)
	}

	eth, err := net.InterfaceByName("enp0s31f6")
	if err != nil {
		panic(err)
	}

	ifis := []*net.Interface{lo, eth}

	sender, err := racket.New(ifis, multicastPool)
	if err != nil {
		panic(err)
	}

	n := 0

	for streamIndex := range 256 {
		st := stream.Stream(fmt.Sprintf("stream-%d", streamIndex))

		for subjectIndex := range 1024 {
			su := subject.Subject{
				Parts: []string{"org", "foo", fmt.Sprintf("racket-%d", subjectIndex)},
			}

			msg := &message.Message{
				Stream:   st,
				Subject:  su,
				Data:     randomBytes(128 + rand.IntN(256)),
				Interval: time.Millisecond*time.Duration(rand.IntN(1000)) + time.Second,
			}

			if err := sender.Publish(msg); err != nil {
				panic(err)
			}

			n++
		}
	}

	log.Printf("Published %d messages\n", n)

	for range 30 {
		stats := sender.Stats()
		spew.Dump("Sender stats\n", stats.Streams[stream.Stream("stream-1")])

		time.Sleep(time.Second)
	}

	sender.Flush()
}
