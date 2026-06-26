package ratelimiter

import (
	"testing"
	"time"
)

func TestRateLimiter_Stop_TerminatesCleanup(t *testing.T) {
	rl := NewRateLimiter(10, 60, 10*time.Millisecond)

	rl.Allow("1.1.1.1")
	rl.Allow("2.2.2.2")

	stopped := make(chan struct{})
	go func() {
		rl.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop() did not return in time — cleanup goroutine may be blocked")
	}
}

func TestRateLimiter_Allow_AfterStop(t *testing.T) {
	rl := NewRateLimiter(5, 60, time.Minute)
	rl.Stop()

	for i := 0; i < 5; i++ {
		ok, remaining, _ := rl.Allow("1.1.1.1")
		if !ok {
			t.Fatalf("Allow() returned false on attempt %d, expected true", i+1)
		}
		if remaining < 0 {
			t.Errorf("remaining = %d, want >= 0", remaining)
		}
	}
}

func TestRateLimiter_Allow_BlocksWhenLimitExceeded(t *testing.T) {
	rl := NewRateLimiter(3, 60, time.Minute)
	defer rl.Stop()

	ip := "10.0.0.1"

	for i := 0; i < 3; i++ {
		ok, _, _ := rl.Allow(ip)
		if !ok {
			t.Fatalf("Allow() returned false on attempt %d, expected true", i+1)
		}
	}

	ok, remaining, _ := rl.Allow(ip)
	if ok {
		t.Error("Allow() returned true after limit exceeded, expected false")
	}
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0 after limit exceeded", remaining)
	}
}
