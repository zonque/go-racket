package main

import (
	"context"
	"fmt"
	"net"

	"github.com/holoplot/go-rotor/pkg/rotor"
	"github.com/holoplot/go-rotor/pkg/rotor/message"
	"github.com/holoplot/go-rotor/pkg/rotor/stream"
	"github.com/holoplot/go-rotor/pkg/rotor/subject"
	"github.com/holoplot/go-rotor/pkg/rotor/subscription"
)

func main() {
	base := net.IPNet{
		IP:   net.ParseIP("239.0.0.0"),
		Mask: net.CIDRMask(16, 32),
	}

	multicastPool := rotor.NewMulticastPool(base)

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

	receiver := rotor.NewReceiver(ifis, multicastPool)

	if _, err := receiver.Subscribe(g1, s1, func(msg *message.Message) {
		fmt.Printf("Received message on subject %s (%d bytes)\n", msg.Subject.String(), len(msg.Data))
	}, subscription.OnlyOnChange()); err != nil {
		panic(err)
	}

	<-context.Background().Done()
}
