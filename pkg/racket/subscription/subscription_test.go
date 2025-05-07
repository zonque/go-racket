package subscription

import (
	"testing"

	"github.com/holoplot/go-racket/pkg/racket/message"
	"github.com/holoplot/go-racket/pkg/racket/subject"
)

func TestTree_Add(t *testing.T) {
	tree := NewTree()
	subj := subject.Subject{Parts: []string{"a", "b", "c"}}
	callback := func(msg *message.Message) {}

	sub := tree.Add(subj, callback)

	if sub == nil {
		t.Fatal("expected subscription to be created, got nil")
	}

	if len(tree.root.children) != 1 {
		t.Errorf("expected 1 child in root, got %d", len(tree.root.children))
	}

	if len(tree.root.children["a"].children["b"].children["c"].subscriptions) != 1 {
		t.Errorf("expected 1 subscription, got %d", len(tree.root.children["a"].children["b"].children["c"].subscriptions))
	}
}

func TestTree_Remove(t *testing.T) {
	tree := NewTree()
	subj := subject.Subject{Parts: []string{"a", "b", "c"}}
	callback := func(msg *message.Message) {}

	sub := tree.Add(subj, callback)
	tree.Remove(sub)

	if len(tree.root.children) != 0 {
		t.Errorf("expected no children in root, got %d", len(tree.root.children))
	}
}

func TestTree_Dispatch(t *testing.T) {
	tree := NewTree()
	subj := subject.Subject{Parts: []string{"a", "b", "c"}}
	called := false
	callback := func(msg *message.Message) {
		called = true
	}

	tree.Add(subj, callback)

	msg := &message.Message{
		Subject: subj,
	}

	tree.Dispatch(msg)

	if !called {
		t.Error("expected callback to be called, but it was not")
	}
}

func TestTree_Dispatch_OnlyOnChange(t *testing.T) {
	tree := NewTree()
	subj := subject.Subject{Parts: []string{"a", "b", "c"}}
	calls := 0
	callback := func(msg *message.Message) {
		calls++
	}

	tree.Add(subj, callback, OnlyOnChange())

	msg := &message.Message{
		Subject: subj,
		Data:    []byte("foo"),
	}

	tree.Dispatch(msg)
	tree.Dispatch(msg)

	if calls != 1 {
		t.Errorf("expected callback to be called once, but it was called %d times", calls)
	}

	msg = &message.Message{
		Subject: subj,
		Data:    []byte("bar"),
	}

	tree.Dispatch(msg)

	if calls != 2 {
		t.Errorf("expected callback to be called twice, but it was called %d times", calls)
	}
}

func TestNode_RemoveSubscription(t *testing.T) {
	root := newNode()
	sub := &Subscription{}
	root.subscriptions = append(root.subscriptions, sub)

	root.removeSubscription(sub)

	if len(root.subscriptions) != 0 {
		t.Errorf("expected no subscriptions, got %d", len(root.subscriptions))
	}
}

func TestNewTree(t *testing.T) {
	tree := NewTree()
	if tree == nil {
		t.Fatal("expected tree to be created, got nil")
	}
	if tree.root == nil {
		t.Fatal("expected root node to be created, got nil")
	}
}
