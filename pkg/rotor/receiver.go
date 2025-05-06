package rotor

import (
	"net"
	"sync"

	"github.com/holoplot/go-rotor/pkg/multicast"
	"github.com/holoplot/go-rotor/pkg/rotor/group"
	"github.com/holoplot/go-rotor/pkg/rotor/message"
	"github.com/holoplot/go-rotor/pkg/rotor/subject"
	"github.com/holoplot/go-rotor/pkg/rotor/subscription"
)

type Receiver struct {
	mutex sync.Mutex

	groups        map[group.Group]receiverGroup
	MulticastPool *MulticastPool
	dispatcher    *multicast.Dispatcher
}

type receiverGroup struct {
	consumer         *multicast.Consumer
	subscriptionTree *subscription.Tree
}

func (r *Receiver) rawReceive(payload []byte) {
	msg, err := message.Parse(payload)
	if err != nil {
		panic(err)
	}

	group, ok := r.groups[msg.Group]
	if !ok {
		panic("group not found")
	}

	group.subscriptionTree.Dispatch(msg)
}

func (r *Receiver) Subscribe(group group.Group, subject subject.Subject, cb subscription.Callback, opts ...subscription.Opt) (*subscription.Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	addr := r.MulticastPool.AddressForGroup(group)

	g, ok := r.groups[group]
	if !ok {
		g = receiverGroup{
			subscriptionTree: subscription.NewTree(),
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

func (r *Receiver) Unsubscribe(group group.Group, sub *subscription.Subscription) error {
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

	r.groups = make(map[group.Group]receiverGroup)
}

func NewReceiver(ifis []*net.Interface, pool *MulticastPool) *Receiver {
	return &Receiver{
		groups:        make(map[group.Group]receiverGroup),
		dispatcher:    multicast.NewDispatcher(ifis),
		MulticastPool: pool,
	}
}
