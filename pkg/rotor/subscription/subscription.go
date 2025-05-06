package subscription

import (
	"crypto/sha512"
	"fmt"
	"sync"

	"github.com/holoplot/go-rotor/pkg/rotor/message"
	"github.com/holoplot/go-rotor/pkg/rotor/subject"
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

	for _, child := range sn.children {
		child.removeSubscription(sub)
	}
}

type Tree struct {
	mutex sync.Mutex
	root  *node
}

func (t *Tree) Add(subject subject.Subject, callback Callback, opts ...Opt) *Subscription {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	node := t.root
	for _, part := range subject.Parts {
		if part == "*" {
			break
		}

		if _, ok := node.children[part]; !ok {
			node.children[part] = newNode()
		}

		node = node.children[part]
	}

	sub := Subscription{
		cb:          callback,
		contentHash: make(map[string]string),
	}

	for _, opt := range opts {
		opt.apply(&sub)
	}

	node.subscriptions = append(node.subscriptions, &sub)

	return &sub
}

func (t *Tree) Remove(sub *Subscription) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.root.removeSubscription(sub)
}

func (t *Tree) Dispatch(msg *message.Message) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var hash string

	node := t.root
	for _, part := range msg.Subject.Parts {
		for _, sub := range node.subscriptions {
			if len(hash) == 0 {
				h := sha512.New()
				h.Write([]byte(msg.Data))
				hash = string(h.Sum(nil))
			}

			if sub.onlyOnChange && sub.contentHash[msg.Subject.String()] == hash {
				continue
			}

			// fmt.Printf("Subject: %s, Hash: %s, OnlyOnChange %t\n", msg.Subject.String(), hash, sub.onlyOnChange)

			sub.contentHash[msg.Subject.String()] = hash

			sub.cb(msg)
		}

		if child, ok := node.children[part]; ok {
			node = child
		} else {
			return
		}
	}
}

func (st *Tree) Dump() string {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	var dump string
	var dumpNode func(node *node, level int)

	dumpNode = func(node *node, level int) {
		prefix := make([]byte, level)
		for i := range prefix {
			prefix[i] = ' '
		}

		dump += string(prefix) + fmt.Sprintf("%d Subscription\n", len(node.subscriptions))

		for part, child := range node.children {
			dump += string(prefix) + part + "\n"
			dumpNode(child, level+4)
		}
	}

	dumpNode(st.root, 0)

	return dump
}

func NewTree() *Tree {
	return &Tree{
		root: newNode(),
	}
}
