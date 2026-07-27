package ratelimiter

import (
	"testing"
	"time"
)

func TestWindow_Allow_BelowLimit(t *testing.T) {
	w := NewWindow(5, 60*time.Second)

	ok, remaining, _ := w.Allow("10.0.0.1")
	if !ok {
		t.Fatal("Allow() = false, want true")
	}
	if remaining != 4 {
		t.Errorf("remaining = %d, want 4", remaining)
	}
}

func TestWindow_Allow_AtLimit(t *testing.T) {
	w := NewWindow(3, 60*time.Second)

	for i := range 3 {
		ok, _, _ := w.Allow("10.0.0.1")
		if !ok {
			t.Fatalf("Allow() = false on attempt %d, want true", i+1)
		}
	}

	ok, remaining, _ := w.Allow("10.0.0.1")
	if ok {
		t.Error("Allow() = true after limit, want false")
	}
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0", remaining)
	}
}

func TestWindow_Allow_DifferentIPs(t *testing.T) {
	w := NewWindow(2, 60*time.Second)

	ok, _, _ := w.Allow("10.0.0.1")
	if !ok {
		t.Fatal("first IP blocked prematurely")
	}

	ok, _, _ = w.Allow("10.0.0.2")
	if !ok {
		t.Fatal("second IP blocked by first IP's count")
	}

	ok, _, _ = w.Allow("10.0.0.1")
	if !ok {
		t.Fatal("first IP blocked on second request")
	}

	ok, _, _ = w.Allow("10.0.0.1")
	if ok {
		t.Error("first IP should be blocked now (3 requests, limit 2)")
	}
}

func TestWindow_Allow_WindowExpiry(t *testing.T) {
	w := NewWindow(2, 1*time.Second)

	for i := range 2 {
		ok, _, _ := w.Allow("10.0.0.1")
		if !ok {
			t.Fatalf("Allow() = false on attempt %d, want true", i+1)
		}
	}

	ok, _, _ := w.Allow("10.0.0.1")
	if ok {
		t.Fatal("should be blocked at limit")
	}

	time.Sleep(1100 * time.Millisecond)

	ok, _, _ = w.Allow("10.0.0.1")
	if !ok {
		t.Error("Allow() = false after window expiry, want true")
	}
}
