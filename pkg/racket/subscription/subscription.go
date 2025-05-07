package subscription

import (
	"sync"

	"github.com/holoplot/go-racket/pkg/racket/message"
	"github.com/holoplot/go-racket/pkg/racket/subject"
)

type Callback func(*message.Message)

type Opt interface {
	apply(*Subscription)
}

type OptOnlyOnChange struct{}

func OnlyOnChange() Opt {
	return &OptOnlyOnChange{}
}

func (s *OptOnlyOnChange) apply(sub *Subscription) {
	sub.onlyOnChange = true
}

type Subscription struct {
	cb           Callback
	onlyOnChange bool
	contentHash  map[string]string
}

type node struct {
	subscriptions []*Subscription
	children      map[string]*node
}

func (n *node) Dispatch(msg *message.Message) {
	for _, sub := range n.subscriptions {
		if sub.onlyOnChange {
			if sub.contentHash[msg.Subject.String()] == msg.Hash() {
				continue
			}

			sub.contentHash[msg.Subject.String()] = msg.Hash()
		}

		sub.cb(msg)
	}
}

func newNode() *node {
	return &node{
		subscriptions: make([]*Subscription, 0),
		children:      make(map[string]*node),
	}
}

func (sn *node) removeSubscription(sub *Subscription) {
	for i, s := range sn.subscriptions {
		if s == sub {
			sn.subscriptions = append(sn.subscriptions[:i], sn.subscriptions[i+1:]...)
			break
		}
	}

	for name, child := range sn.children {
		child.removeSubscription(sub)

		if len(child.subscriptions) == 0 && len(child.children) == 0 {
			delete(sn.children, name)
		}
	}
}

type Tree struct {
	mutex sync.Mutex
	root  *node
}

func (t *Tree) Add(s subject.Subject, callback Callback, opts ...Opt) *Subscription {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	node := t.root
	for _, part := range s.Parts {
		if part == subject.Wildcard {
			break
		}

		if _, ok := node.children[part]; !ok {
			node.children[part] = newNode()
		}

		node = node.children[part]
	}

	sub := &Subscription{
		cb:          callback,
		contentHash: make(map[string]string),
	}

	for _, opt := range opts {
		opt.apply(sub)
	}

	node.subscriptions = append(node.subscriptions, sub)

	return sub
}

func (t *Tree) Remove(sub *Subscription) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.root.removeSubscription(sub)
}

func (t *Tree) Dispatch(msg *message.Message) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	node := t.root
	for _, part := range msg.Subject.Parts {
		node.Dispatch(msg)

		if child, ok := node.children[part]; ok {
			node = child
		} else {
			// We have reached the end of the subject parts, and
			// there are no more children to traverse.
			return
		}
	}

	node.Dispatch(msg)
}

func NewTree() *Tree {
	return &Tree{
		root: newNode(),
	}
}
