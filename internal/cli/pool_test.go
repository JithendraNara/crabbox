package cli

import (
	"testing"
	"time"
)

func TestShouldCleanupServerSkipsActiveStates(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	for _, state := range []string{"leased", "ready", "running", "active", "provisioning"} {
		server := Server{Labels: map[string]string{
			"keep":       "false",
			"state":      state,
			"expires_at": now.Add(-time.Hour).Format(time.RFC3339),
		}}
		if ok, reason := shouldCleanupServer(server, now); ok {
			t.Fatalf("shouldCleanupServer state=%s=%v, %s; want skip", state, ok, reason)
		}
	}
}

func TestShouldCleanupServerDeletesStaleActiveStates(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	server := Server{Labels: map[string]string{
		"keep":       "false",
		"state":      "ready",
		"expires_at": now.Add(-13 * time.Hour).Format(time.RFC3339),
	}}
	if ok, reason := shouldCleanupServer(server, now); !ok {
		t.Fatalf("shouldCleanupServer=%v, %s; want delete", ok, reason)
	}
}

func TestShouldCleanupServerDeletesExpiredInactive(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	server := Server{Labels: map[string]string{
		"keep":       "false",
		"expires_at": now.Add(-time.Minute).Format(time.RFC3339),
	}}
	if ok, reason := shouldCleanupServer(server, now); !ok {
		t.Fatalf("shouldCleanupServer=%v, %s; want delete", ok, reason)
	}
}

func TestShouldCleanupServerKeepsUnexpiredAndKept(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	tests := []Server{
		{Labels: map[string]string{"keep": "true", "expires_at": now.Add(-time.Hour).Format(time.RFC3339)}},
		{Labels: map[string]string{"keep": "false", "expires_at": now.Add(time.Hour).Format(time.RFC3339)}},
		{Labels: map[string]string{"keep": "false"}},
	}
	for _, server := range tests {
		if ok, reason := shouldCleanupServer(server, now); ok {
			t.Fatalf("shouldCleanupServer=%v, %s; want skip", ok, reason)
		}
	}
}

func TestHeartbeatInterval(t *testing.T) {
	tests := map[time.Duration]time.Duration{
		0:                time.Minute,
		9 * time.Second:  5 * time.Second,
		30 * time.Second: 10 * time.Second,
		90 * time.Minute: time.Minute,
	}
	for ttl, want := range tests {
		if got := heartbeatInterval(ttl); got != want {
			t.Fatalf("heartbeatInterval(%s)=%s want %s", ttl, got, want)
		}
	}
}
