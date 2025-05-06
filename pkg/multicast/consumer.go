package multicast

import "net"

type consumers []*Consumer

type Consumer struct {
	addr       *net.UDPAddr
	cb         func([]byte)
	dispatcher *Dispatcher
}

func (c *Consumer) Close() {
	c.dispatcher.removeConsumer(c)
}
