package main

import (
	"context"
	"fmt"
	"net"

	"github.com/holoplot/go-rotor/pkg/rotor"
)

func main() {
	base := net.IPNet{
		IP:   net.ParseIP("239.0.0.0"),
		Mask: net.CIDRMask(16, 32),
	}

	multicastPool := rotor.NewMulticastPool(base)

	g1 := rotor.Group("group-1")

	s1 := rotor.Subject{
		Parts: []string{"org", "holoplot", "go", "*"},
	}

	lo, err := net.InterfaceByName("lo")
	if err != nil {
		panic(err)
	}

	ifis := []*net.Interface{lo}

	receiver := rotor.NewReceiver(ifis, multicastPool)
	if _, err := receiver.Subscribe(g1, s1, func(msg *rotor.Message) {
		fmt.Printf("Received message on subject %s: >%s<\n", msg.Subject.String(), string(msg.Data))
	}); err != nil {
		panic(err)
	}

	<-context.Background().Done()
}
