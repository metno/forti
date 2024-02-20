package server

import (
	"testing"
	"time"
)

func TestExpiresWithOffset(t *testing.T) {
	maxOffset := 180
	now := time.Now()
	expiry := expires(now, maxOffset)
	maxExpiry := now.Add(time.Duration(maxOffset) * time.Second)

	if expiry.Before(now) || expiry.After(maxExpiry) {
		t.Fatalf(`Expected expires(%d) to return value between %v and %v, got %v`,
			maxOffset, now, maxExpiry, expiry)
	}
}

func TestExpiresWithNoOffset(t *testing.T) {
	maxOffset := 0
	now := time.Now()
	expiry := expires(now, maxOffset)
	maxExpiry := now.Add(time.Second * 60)

	if !expiry.Equal(maxExpiry) {
		t.Fatalf(`Expected expires(%d) to return value between %v and %v, got %v`,
			maxOffset, now, maxExpiry, expiry)
	}
}
