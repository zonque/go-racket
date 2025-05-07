package multicastpool

import (
	"net"
	"testing"

	"github.com/holoplot/go-racket/pkg/racket/stream"
)

func TestNew(t *testing.T) {
	_, ipNet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("failed to parse CIDR: %v", err)
	}

	pool := New(*ipNet)
	if pool == nil {
		t.Fatal("expected pool to be created, got nil")
	}

	if pool.base.String() != ipNet.String() {
		t.Errorf("expected base to be %v, got %v", ipNet, pool.base)
	}
}

func TestPool_AddressForStream(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		stream stream.Stream
		want   string
	}{
		{
			name:   "Test 1",
			base:   "239.1.0.0/16",
			stream: stream.Stream("stream-1"),
			want:   "239.1.137.50",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ipNet, err := net.ParseCIDR(tt.base)
			if err != nil {
				t.Fatalf("failed to parse CIDR: %v", err)
			}

			pool := New(*ipNet)
			if pool == nil {
				t.Fatal("expected pool to be created, got nil")
			}

			if got := pool.AddressForStream(tt.stream); got.IP.String() != tt.want {
				t.Errorf("Pool.AddressForStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
