package config

import (
	"testing"
)

// TestManager_GetProfile_ReturnsNamedConfig is the S-30.1 acceptance
// scenario for the new ConfigManager.GetProfile port method. The
// settings-view's persistent profile sidebar previews a non-active
// profile by calling GetProfile(name), so the manager must surface
// the named profile's *Config.
//
// Before the change the method does not exist and the test fails at
// compile time. After the change the method must:
//   - return the matching Profile's *Config for a known name
//   - return nil for an unknown name (no panic)
func TestManager_GetProfile_ReturnsNamedConfig(t *testing.T) {
	pm := newProfileManager(NewLoader())
	pm.profiles = []Profile{
		{
			Name: "default",
			Config: &Config{
				Request: RequestConfig{Method: "GET", URLTemplate: "https://default.example"},
				CSV:     CSVConfig{Fields: []string{"a"}},
				Workers: 1,
			},
		},
		{
			Name: "production",
			Config: &Config{
				Request: RequestConfig{Method: "POST", URLTemplate: "https://prod.example/api"},
				CSV:     CSVConfig{Fields: []string{"id", "name"}},
				Workers: 4,
			},
		},
	}
	pm.activeIndex = 0

	mgr := &managerImpl{
		profileMgr: pm,
		listeners:  make([]func(*Config), 0),
	}

	got := mgr.GetProfile("production")
	if got == nil {
		t.Fatalf("GetProfile(\"production\") returned nil; expected production config")
	}
	if got.Request.URLTemplate != "https://prod.example/api" {
		t.Fatalf("GetProfile(\"production\") returned the wrong config: URLTemplate=%q", got.Request.URLTemplate)
	}
	if got.Request.Method != "POST" {
		t.Fatalf("GetProfile(\"production\") returned the wrong config: Method=%q", got.Request.Method)
	}

	if unknown := mgr.GetProfile("nonexistent"); unknown != nil {
		t.Fatalf("GetProfile(\"nonexistent\") must return nil for unknown profiles, got %+v", unknown)
	}
}

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
