package server

import (
	"testing"
	"time"
)

func TestExpires(t *testing.T) {
	maxOffset := 180
	now := time.Now()
	expiry := expires(now, maxOffset)
	maxExpiry := now.Add(time.Duration(maxOffset) * time.Second)

	if expiry.Before(now) || expiry.After(maxExpiry) {
		t.Fatalf(`Expected expires(%d) to return value between %v and %v, got %v`,
			maxOffset, now, maxExpiry, expiry)
	}

	maxOffset = 0
	now = time.Now()
	expiry = expires(now, maxOffset)
	maxExpiry = now.Add(time.Second)
	if !expiry.Equal(maxExpiry) {
		t.Fatalf(`Expected expires(%d) to return value between %v and %v, got %v`,
			maxOffset, now, maxExpiry, expiry)
	}
}
