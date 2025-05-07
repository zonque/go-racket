package racket

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/holoplot/go-racket/pkg/multicast"
	"github.com/holoplot/go-racket/pkg/racket/message"
	multicastpool "github.com/holoplot/go-racket/pkg/racket/multicast-pool"
	"github.com/holoplot/go-racket/pkg/racket/stream"
	"github.com/holoplot/go-racket/pkg/racket/subject"
	"github.com/holoplot/go-racket/pkg/racket/subscription"
)

type Receiver struct {
	mutex sync.Mutex

	streams       map[stream.Stream]*receiverStream
	MulticastPool *multicastpool.Pool
	dispatcher    *multicast.Dispatcher
}

type receiverStream struct {
	consumer         *multicast.Consumer
	subscriptionTree *subscription.Tree

	messagesReceived   atomic.Uint64
	messagesDispatched atomic.Uint64
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

	d := stream.subscriptionTree.Dispatch(msg)
	stream.messagesReceived.Add(1)
	stream.messagesDispatched.Add(d)
}

func (r *Receiver) Subscribe(stream stream.Stream, subject subject.Subject, cb subscription.Callback, opts ...subscription.Opt) (*subscription.Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	addr := r.MulticastPool.AddressForStream(stream)

	rs, ok := r.streams[stream]
	if !ok {
		rs = &receiverStream{
			subscriptionTree: subscription.NewTree(),
		}

		var err error

		rs.consumer, err = r.dispatcher.AddConsumer(addr, r.rawReceive)
		if err != nil {
			return nil, err
		}

		r.streams[stream] = rs
	}

	sub := rs.subscriptionTree.Add(subject, cb, opts...)

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

	r.streams = make(map[stream.Stream]*receiverStream)
}

func New(ifis []*net.Interface, pool *multicastpool.Pool) *Receiver {
	return &Receiver{
		streams:       make(map[stream.Stream]*receiverStream),
		dispatcher:    multicast.NewDispatcher(ifis),
		MulticastPool: pool,
	}
}

type StreamStats struct {
	SubscriptionStats  subscription.Stats
	MessagesReceived   uint64
	MessagesDispatched uint64
}

type Stats struct {
	Streams map[stream.Stream]StreamStats
}

func (r *Receiver) Stats() Stats {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	stats := Stats{
		Streams: make(map[stream.Stream]StreamStats),
	}

	for stream, g := range r.streams {
		stats.Streams[stream] = StreamStats{
			SubscriptionStats:  g.subscriptionTree.Stats(),
			MessagesReceived:   g.messagesReceived.Load(),
			MessagesDispatched: g.messagesDispatched.Load(),
		}
	}

	return stats
}
