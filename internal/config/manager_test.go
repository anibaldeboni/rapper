package config

import (
	"testing"
)

// TestManager_SetActiveProfile_NotifiesListeners proves that switching the
// active profile fires the same OnChange notification block that Update does.
// Before the fix the listener is never invoked and the test fails.
func TestManager_SetActiveProfile_NotifiesListeners(t *testing.T) {
	pm := newProfileManager(NewLoader())
	pm.profiles = []Profile{
		{
			Name: "p1",
			Config: &Config{
				Request: RequestConfig{Method: "POST", URLTemplate: "https://api1.example", BodyTemplate: "{}"},
				CSV:     CSVConfig{Fields: []string{"a"}},
				Workers: 1,
			},
		},
		{
			Name: "p2",
			Config: &Config{
				Request: RequestConfig{Method: "POST", URLTemplate: "https://api2.example", BodyTemplate: "{}"},
				CSV:     CSVConfig{Fields: []string{"b"}},
				Workers: 1,
			},
		},
	}
	pm.activeIndex = 0

	mgr := &managerImpl{
		profileMgr: pm,
		listeners:  make([]func(*Config), 0),
	}

	var (
		calls int
		got   *Config
	)
	mgr.OnChange(func(cfg *Config) {
		calls++
		got = cfg
	})

	if err := mgr.SetActiveProfile("p2"); err != nil {
		t.Fatalf("SetActiveProfile: %v", err)
	}

	if calls != 1 {
		t.Fatalf("expected OnChange listener to be called exactly once, got %d", calls)
	}
	if got == nil {
		t.Fatalf("OnChange listener received nil config")
	}
	if got.Request.URLTemplate != "https://api2.example" {
		t.Fatalf("expected URLTemplate from profile p2, got %q", got.Request.URLTemplate)
	}
}
