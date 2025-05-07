package racket

import (
	"net"
	"sync"

	"github.com/holoplot/go-racket/pkg/multicast"
	"github.com/holoplot/go-racket/pkg/racket/message"
	"github.com/holoplot/go-racket/pkg/racket/stream"
	"github.com/holoplot/go-racket/pkg/racket/subject"
	"github.com/holoplot/go-racket/pkg/racket/subscription"
)

type Receiver struct {
	mutex sync.Mutex

	streams       map[stream.Stream]receiverStream
	MulticastPool *MulticastPool
	dispatcher    *multicast.Dispatcher
}

type receiverStream struct {
	consumer         *multicast.Consumer
	subscriptionTree *subscription.Tree
}

func (r *Receiver) rawReceive(payload []byte) {
	msg, err := message.Parse(payload)
	if err != nil {
		panic(err)
	}

	stream, ok := r.streams[msg.Stream]
	if !ok {
		panic("stream not found")
	}

	stream.subscriptionTree.Dispatch(msg)
}

func (r *Receiver) Subscribe(stream stream.Stream, subject subject.Subject, cb subscription.Callback, opts ...subscription.Opt) (*subscription.Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	addr := r.MulticastPool.AddressForStream(stream)

	g, ok := r.streams[stream]
	if !ok {
		g = receiverStream{
			subscriptionTree: subscription.NewTree(),
		}

		var err error

		g.consumer, err = r.dispatcher.AddConsumer(addr, r.rawReceive)
		if err != nil {
			return nil, err
		}

		r.streams[stream] = g
	}

	sub := g.subscriptionTree.Add(subject, cb, opts...)

	return sub, nil
}

func (r *Receiver) Unsubscribe(stream stream.Stream, sub *subscription.Subscription) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	g, ok := r.streams[stream]
	if !ok {
		return nil
	}

	g.subscriptionTree.Remove(sub)

	return nil
}

func (r *Receiver) Close() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, g := range r.streams {
		g.consumer.Close()
	}

	r.dispatcher.Close()

	r.streams = make(map[stream.Stream]receiverStream)
}

func NewReceiver(ifis []*net.Interface, pool *MulticastPool) *Receiver {
	return &Receiver{
		streams:       make(map[stream.Stream]receiverStream),
		dispatcher:    multicast.NewDispatcher(ifis),
		MulticastPool: pool,
	}
}
