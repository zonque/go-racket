package main

import (
	"fmt"
	"net"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/holoplot/go-racket/pkg/racket/message"
	multicastpool "github.com/holoplot/go-racket/pkg/racket/multicast-pool"
	racket "github.com/holoplot/go-racket/pkg/racket/receiver"
	"github.com/holoplot/go-racket/pkg/racket/stream"
	"github.com/holoplot/go-racket/pkg/racket/subject"
	"github.com/holoplot/go-racket/pkg/racket/subscription"
)

func main() {
	_, base, _ := net.ParseCIDR("239.0.0.0/16")
	multicastPool, err := multicastpool.New(*base)
	if err != nil {
		panic(err)
	}

	st := stream.Stream("stream-1")
	su := subject.Subject{
		Parts: []string{"org", "foo", "*"},
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

	receiver := racket.New(ifis, multicastPool)

	if _, err := receiver.Subscribe(st, su, func(msg *message.Message) {
		fmt.Printf("Received message on subject %s (%d bytes)\n", msg.Subject, len(msg.Data))
	}, subscription.OnlyOnChange()); err != nil {
		panic(err)
	}

	for {
		stats := receiver.Stats()
		spew.Dump(stats)

		time.Sleep(time.Second)
	}
}
