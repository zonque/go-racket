package main

import (
	"context"
	"fmt"
	"net"

	"github.com/holoplot/go-racket/pkg/racket/message"
	multicastpool "github.com/holoplot/go-racket/pkg/racket/multicast-pool"
	racket "github.com/holoplot/go-racket/pkg/racket/receiver"
	"github.com/holoplot/go-racket/pkg/racket/stream"
	"github.com/holoplot/go-racket/pkg/racket/subject"
	"github.com/holoplot/go-racket/pkg/racket/subscription"
)

func main() {
	_, base, _ := net.ParseCIDR("239.0.0.0/16")
	multicastPool := multicastpool.New(*base)

	g1 := stream.Stream("stream-1")

	s1 := subject.Subject{
		Parts: []string{"org", "holoplot", "*"},
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

	if _, err := receiver.Subscribe(g1, s1, func(msg *message.Message) {
		fmt.Printf("Received message on subject %s (%d bytes)\n", msg.Subject.String(), len(msg.Data))
	}, subscription.OnlyOnChange()); err != nil {
		panic(err)
	}

	<-context.Background().Done()
}
