package main

import "testing"

func TestServiceStatusTransitions(t *testing.T) {
	app := &App{}
	s := NewService(app)

	if s.status != StatusIdle {
		t.Errorf("expected initial status idle, got %s", s.status)
	}

	// Start should succeed when config is available (found at ~/.cc-connect/config.toml)
	err := s.Start()
	if err != nil {
		t.Errorf("unexpected Start error: %v", err)
	}
	if s.status != StatusRunning {
		t.Errorf("expected status running after Start, got %s", s.status)
	}

	// Stop should succeed when running
	err = s.Stop()
	if err != nil {
		t.Errorf("unexpected Stop error: %v", err)
	}
	if s.status != StatusIdle {
		t.Errorf("expected status idle after Stop, got %s", s.status)
	}

	// Stop when already idle returns nil (no-op)
	err = s.Stop()
	if err != nil {
		t.Errorf("expected Stop to return nil when not running, got %v", err)
	}
}

func TestServiceStatusConstants(t *testing.T) {
	tests := map[ServiceStatus]string{
		StatusIdle:     "idle",
		StatusStarting: "starting",
		StatusRunning:  "running",
		StatusStopping: "stopping",
		StatusError:    "error",
	}
	for status, expected := range tests {
		if string(status) != expected {
			t.Errorf("expected %s, got %s", expected, string(status))
		}
	}
}
