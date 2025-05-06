package rotor

import (
	"fmt"
	"sync"
)

type Callback func(*Message)

type Subscription struct {
	cb Callback
}

type subscriptionNode struct {
	subscriptions []*Subscription
	children      map[string]*subscriptionNode
}

func newSubscriptionNode() *subscriptionNode {
	return &subscriptionNode{
		subscriptions: make([]*Subscription, 0),
		children:      make(map[string]*subscriptionNode),
	}
}

func (sn *subscriptionNode) removeSubscription(sub *Subscription) {
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

type SubscriptionTree struct {
	mutex sync.RWMutex
	root  *subscriptionNode
}

func (st *SubscriptionTree) Add(subject Subject, callback Callback) *Subscription {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	node := st.root
	for _, part := range subject.Parts {
		if part == "*" {
			break
		}

		if _, ok := node.children[part]; !ok {
			node.children[part] = newSubscriptionNode()
		}

		node = node.children[part]
	}

	sub := Subscription{cb: callback}

	node.subscriptions = append(node.subscriptions, &sub)

	return &sub
}

func (st *SubscriptionTree) Remove(sub *Subscription) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.root.removeSubscription(sub)
}

func (st *SubscriptionTree) Call(msg *Message) {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	node := st.root
	for _, part := range msg.Subject.Parts {
		for _, sub := range node.subscriptions {
			sub.cb(msg)
		}

		if child, ok := node.children[part]; ok {
			node = child
		} else {
			return
		}
	}
}

func (st *SubscriptionTree) Dump() string {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	var dump string
	var dumpNode func(node *subscriptionNode, level int)

	dumpNode = func(node *subscriptionNode, level int) {
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

func NewSubscriptionTree() *SubscriptionTree {
	return &SubscriptionTree{
		root: newSubscriptionNode(),
	}
}
