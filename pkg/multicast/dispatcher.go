package multicast

import (
	"net"
	"sync"
)

type Dispatcher struct {
	mutex     sync.Mutex
	ifis      []*net.Interface
	listeners map[int]*listener
}

func (d *Dispatcher) AddConsumer(addr *net.UDPAddr, cb func([]byte)) (*Consumer, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	c := &Consumer{
		addr:       addr,
		cb:         cb,
		dispatcher: d,
	}

	l, ok := d.listeners[addr.Port]
	if !ok {
		var err error

		l, err = newListener(addr.Port, d.ifis)
		if err != nil {
			return nil, err
		}

		d.listeners[addr.Port] = l
	}

	if err := l.addConsumer(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (d *Dispatcher) removeConsumer(c *Consumer) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	l, ok := d.listeners[c.addr.Port]
	if !ok {
		return
	}

	l.removeConsumer(c)

	if !l.hasConsumers() {
		l.close()

		delete(d.listeners, c.addr.Port)
	}
}

func (d *Dispatcher) Close() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	for _, l := range d.listeners {
		l.close()
	}

	clear(d.listeners)
}

func (d *Dispatcher) Interfaces() []*net.Interface {
	return d.ifis
}

func NewDispatcher(ifis []*net.Interface) *Dispatcher {
	return &Dispatcher{
		ifis:      ifis,
		listeners: make(map[int]*listener),
	}
}
