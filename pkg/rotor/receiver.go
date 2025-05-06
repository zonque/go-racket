package rotor

import (
	"net"
	"sync"

	"github.com/holoplot/go-rotor/pkg/multicast"
)

type Receiver struct {
	mutex sync.Mutex

	groups map[Group]receiverGroup

	MulticastPool *MulticastPool
	dispatcher    *multicast.Dispatcher
}

type receiverGroup struct {
	consumer         *multicast.Consumer
	subscriptionTree *SubscriptionTree
}

func (r *Receiver) rawReceive(payload []byte) {
	msg, err := ParseMessage(payload)
	if err != nil {
		panic(err)
	}

	group, ok := r.groups[msg.Group]
	if !ok {
		panic("group not found")
	}

	group.subscriptionTree.Call(msg)
}

func (r *Receiver) Subscribe(group Group, subject Subject, cb Callback, opts ...SubscriptionOpt) (*Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	addr := r.MulticastPool.AddressForGroup(group)

	g, ok := r.groups[group]
	if !ok {
		g = receiverGroup{
			subscriptionTree: NewSubscriptionTree(),
		}

		var err error

		g.consumer, err = r.dispatcher.AddConsumer(addr, r.rawReceive)
		if err != nil {
			return nil, err
		}

		r.groups[group] = g
	}

	sub := g.subscriptionTree.Add(subject, cb, opts...)

	return sub, nil
}

func (r *Receiver) Unsubscribe(group Group, sub *Subscription) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	g, ok := r.groups[group]
	if !ok {
		return nil
	}

	g.subscriptionTree.Remove(sub)

	return nil
}

func (r *Receiver) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, g := range r.groups {
		g.consumer.Close()
	}

	r.dispatcher.Close()

	r.groups = make(map[Group]receiverGroup)
}

func NewReceiver(ifis []*net.Interface, pool *MulticastPool) *Receiver {
	return &Receiver{
		groups:        make(map[Group]receiverGroup),
		dispatcher:    multicast.NewDispatcher(ifis),
		MulticastPool: pool,
	}
}
