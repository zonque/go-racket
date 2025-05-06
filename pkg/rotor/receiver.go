package rotor

import (
	"net"
	"sync"

	"github.com/holoplot/go-rotor/pkg/multicast"
)

type Receiver struct {
	mutex sync.Mutex

	subscriptions map[string]subscription

	MulticastPool *MulticastPool
	dispatcher    *multicast.Dispatcher
}

type subscription struct {
	cb       Callback
	consumer *multicast.Consumer
}

type Callback func(*Message)

func (r *Receiver) rawReceive(payload []byte) {
	msg, err := ParseMessage(payload)
	if err != nil {
		panic(err)
	}

	sub, ok := r.subscriptions[msg.Subject.String()]
	if !ok {
		return
	}

	sub.cb(msg)
}

func (r *Receiver) Subscribe(subject Subject, cb Callback) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	addr := r.MulticastPool.AddressForSubject(subject)

	consumer, err := r.dispatcher.AddConsumer(addr, r.rawReceive)
	if err != nil {
		return err
	}

	sub := subscription{
		cb:       cb,
		consumer: consumer,
	}

	r.subscriptions[subject.String()] = sub

	return nil
}

func (r *Receiver) Unsubscribe(subject Subject) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	sub, ok := r.subscriptions[subject.String()]
	if !ok {
		return nil
	}

	sub.consumer.Close()

	delete(r.subscriptions, subject.String())

	return nil
}

func (r *Receiver) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, sub := range r.subscriptions {
		sub.consumer.Close()
	}

	r.dispatcher.Close()

	r.subscriptions = make(map[string]subscription)
}

func NewReceiver(ifis []*net.Interface, pool *MulticastPool) *Receiver {
	return &Receiver{
		subscriptions: make(map[string]subscription),
		dispatcher:    multicast.NewDispatcher(ifis),
		MulticastPool: pool,
	}
}
