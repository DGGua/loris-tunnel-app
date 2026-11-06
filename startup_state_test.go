package main

import (
	"testing"
	"time"

	"loris-tunnel/internal/conf"
	"loris-tunnel/internal/model"
)

func TestGetStateDoesNotWaitForConfigGate(t *testing.T) {
	a, _ := syncTestApp(t)
	_, err := a.storage.Update(func(c *conf.Config) error {
		c.Tunnels = []model.Tunnel{{ID: 1, Status: "running", ManuallyStopped: true}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	a.configGate.Lock()
	defer a.configGate.Unlock()
	type result struct {
		state model.State
		err   error
	}
	ready := make(chan result, 1)
	go func() { state, err := a.GetState(); ready <- result{state, err} }()
	select {
	case got := <-ready:
		if got.err != nil {
			t.Fatal(got.err)
		}
		if len(got.state.Jumpers) != 1 || len(got.state.Groups) != 1 || len(got.state.Tunnels) != 1 {
			t.Fatal("incomplete configuration snapshot")
		}
		if got.state.Tunnels[0].Status != "stopped" || !got.state.Tunnels[0].ManuallyStopped {
			t.Fatal("snapshot did not overlay runtime state")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("page data waited for a network operation's gate")
	}
}

func TestGetStateReportsUnlimitedTunnelLimit(t *testing.T) {
	a, _ := syncTestApp(t)
	state, err := a.GetState()
	if err != nil {
		t.Fatal(err)
	}
	if state.ActiveSlotLimit != 0 {
		t.Fatalf("ActiveSlotLimit = %d, expected 0 (unlimited)", state.ActiveSlotLimit)
	}
}
