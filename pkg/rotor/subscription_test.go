package rotor

import (
	"reflect"
	"testing"
)

func TestSubscriptionTree_Add(t *testing.T) {
	tests := []struct {
		name     string
		subject  Subject
		callback Callback
		want     *SubscriptionTree
	}{
		{
			name:    "Test adding a subscription",
			subject: Subject{Parts: []string{"test", "subject"}},
			want: &SubscriptionTree{
				root: &subscriptionNode{
					subscriptions: []*Subscription{},
					children: map[string]*subscriptionNode{
						"test": {
							subscriptions: []*Subscription{},
							children: map[string]*subscriptionNode{
								"subject": {
									subscriptions: []*Subscription{},
									children:      map[string]*subscriptionNode{},
								},
							},
						},
					},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := NewSubscriptionTree()

			st.Add(tt.subject, tt.callback)

			if !reflect.DeepEqual(*st.root, *tt.want.root) {
				t.Errorf("SubscriptionTree.Add() = %+v, want %+v", *st.root, *tt.want.root)
			}
		})
	}
}
