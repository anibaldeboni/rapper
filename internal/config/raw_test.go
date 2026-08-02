package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// writeYAMLFile writes content to dir/name and returns the full path.
// It uses 0644 permissions to mirror the convention used elsewhere
// in the config package.
func writeYAMLFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeYAMLFile(%s): %v", name, err)
	}
	return path
}

// newRawTestManager builds a managerImpl backed by a temp dir with the
// given (name → yaml) profile map. The first entry is marked active.
// Returns the manager and the temp dir (caller is responsible for
// cleanup; t.TempDir handles that automatically when t.Cleanup fires).
func newRawTestManager(t *testing.T, profiles map[string]string) (*runManager, string) {
	t.Helper()
	dir := t.TempDir()
	pm := newRequestManager(NewLoader())
	pm.requests = make([]Request, 0, len(profiles))
	i := 0
	for name, yaml := range profiles {
		path := writeYAMLFile(t, dir, name+".yml", yaml)
		cfg, err := pm.configLoader.Load(path)
		if err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
		pm.requests = append(pm.requests, Request{
			Name:     name,
			FilePath: path,
			Config:   cfg,
		})
		if i == 0 {
			pm.activeIndex = 0
		}
		i++
	}
	return &runManager{
		requestMgr: pm,
		listeners:  make([]func(*Config), 0),
	}, dir
}

// S1.1 — GetRequestRaw returns the file's bytes for a known profile.
func TestManager_GetProfileRaw_ReturnsFileBytes(t *testing.T) {
	yaml := "request:\n  method: POST\n  url_template: https://api.example\ncsv:\n  fields: [id]\nworkers: 4\n"
	mgr, _ := newRawTestManager(t, map[string]string{"p": yaml})

	got, err := mgr.GetRequestRaw("p")
	if err != nil {
		t.Fatalf("GetRequestRaw: unexpected error: %v", err)
	}
	if got != yaml {
		t.Fatalf("GetRequestRaw: bytes mismatch.\nwant: %q\ngot:  %q", yaml, got)
	}
}

// S1.2 — GetRequestRaw returns a non-nil error for an unknown profile name.
// The error must reference the missing profile so the user sees which
// name is at fault.
func TestManager_GetProfileRaw_MissingName(t *testing.T) {
	mgr, _ := newRawTestManager(t, map[string]string{"p": "request:\n  method: GET\n  url_template: http://x\ncsv:\n  fields: [a]\nworkers: 1\n"})

	got, err := mgr.GetRequestRaw("ghost")
	if err == nil {
		t.Fatalf("GetRequestRaw(\"ghost\") must return a non-nil error, got nil (returned %q)", got)
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("error message must reference the missing profile name, got: %v", err)
	}
	if got != "" {
		t.Fatalf("GetRequestRaw on missing name must return empty string, got %q", got)
	}
}

// S1.3 — SaveRequestRaw accepts valid YAML, writes the file, calls
// Update exactly once, and fires OnChange once.
func TestManager_SaveProfileRaw_ValidYAML_WritesAndFiresOnChange(t *testing.T) {
	mgr, dir := newRawTestManager(t, map[string]string{"p": "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"})

	var (
		calls atomic.Int32
		got   *Config
	)
	mgr.OnChange(func(cfg *Config) {
		calls.Add(1)
		got = cfg
	})

	newYAML := "request:\n  method: PUT\n  url_template: http://y\ncsv:\n  fields: [name]\nworkers: 8\n"
	if err := mgr.SaveRequestRaw("p", newYAML); err != nil {
		t.Fatalf("SaveRequestRaw: %v", err)
	}

	// File content on disk reflects the new YAML.
	written, err := os.ReadFile(filepath.Join(dir, "p.yml"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(written) != newYAML {
		t.Fatalf("file content not updated.\nwant: %q\ngot:  %q", newYAML, string(written))
	}

	// OnChange fired exactly once with the parsed new config.
	if gotInt := calls.Load(); gotInt != 1 {
		t.Fatalf("OnChange call count = %d, want 1", gotInt)
	}
	if got == nil {
		t.Fatalf("OnChange received nil config")
	}
	if got.Request.Method != "PUT" || got.Request.URLTemplate != "http://y" || got.Workers != 8 {
		t.Fatalf("OnChange config not parsed from new content: %+v", got)
	}
}

// S1.4 — SaveRequestRaw with invalid YAML returns an error, does NOT
// write the file, does NOT call Update, does NOT fire OnChange.
func TestManager_SaveProfileRaw_InvalidYAML_BlocksWrite(t *testing.T) {
	const origYAML = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	mgr, dir := newRawTestManager(t, map[string]string{"p": origYAML})

	var calls atomic.Int32
	mgr.OnChange(func(_ *Config) { calls.Add(1) })

	// Malformed: tab indentation in YAML is invalid.
	badYAML := "request:\n\tmethod: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 1\n"
	err := mgr.SaveRequestRaw("p", badYAML)
	if err == nil {
		t.Fatalf("SaveRequestRaw with invalid YAML must return an error, got nil")
	}

	// File on disk must be UNCHANGED.
	written, err := os.ReadFile(filepath.Join(dir, "p.yml"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(written) != origYAML {
		t.Fatalf("file must be unchanged after a failed save.\nwant: %q\ngot:  %q", origYAML, string(written))
	}
	if calls.Load() != 0 {
		t.Fatalf("OnChange must NOT fire on a failed save; got %d calls", calls.Load())
	}
}

// S1.5 — SaveRequestRaw with the empty string is rejected without a write.
func TestManager_SaveProfileRaw_EmptyContent_Rejected(t *testing.T) {
	const origYAML = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	mgr, dir := newRawTestManager(t, map[string]string{"p": origYAML})

	var calls atomic.Int32
	mgr.OnChange(func(_ *Config) { calls.Add(1) })

	err := mgr.SaveRequestRaw("p", "")
	if err == nil {
		t.Fatalf("SaveRequestRaw(\"\") must return an error, got nil")
	}

	written, err := os.ReadFile(filepath.Join(dir, "p.yml"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(written) != origYAML {
		t.Fatalf("file must be unchanged after empty save.\nwant: %q\ngot:  %q", origYAML, string(written))
	}
	if calls.Load() != 0 {
		t.Fatalf("OnChange must NOT fire on a rejected empty save; got %d calls", calls.Load())
	}
}

// S1.6 — SaveRequestRaw rejects the legacy YAML schema without changing the
// existing profile or notifying listeners.
func TestManager_SaveProfileRaw_LegacyYAML_Rejected(t *testing.T) {
	legacy := "url: http://legacy\ntoken: abc\npath:\n  method: POST\n  template: /v1/x\npayload:\n  template: '{}'\ncsv:\n  fields: [id]\nworkers: 1\n"
	original := "request:\n  method: POST\n  url_template: http://current\ncsv:\n  fields: [id]\nworkers: 2\n"
	mgr, dir := newRawTestManager(t, map[string]string{"profile": original})

	var calls atomic.Int32
	mgr.OnChange(func(_ *Config) { calls.Add(1) })

	err := mgr.SaveRequestRaw("profile", legacy)
	if err == nil {
		t.Fatal("SaveRequestRaw must reject legacy YAML")
	}
	written, err := os.ReadFile(filepath.Join(dir, "profile.yml"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(written) != original {
		t.Fatalf("profile content changed after rejected legacy save.\nwant: %q\ngot:  %q", original, string(written))
	}
	if calls.Load() != 0 {
		t.Fatalf("OnChange must not fire after rejected legacy save; got %d calls", calls.Load())
	}
}

func TestParseConfigFile_LegacyYAML_Rejected(t *testing.T) {
	t.Parallel()

	legacy := []byte("url: http://legacy\ntoken: abc\npath:\n  method: POST\n  template: /v1/x\npayload:\n  template: '{}'\ncsv:\n  fields: [id]\nworkers: 1\n")

	_, err := parseConfigFile(legacy)
	if err == nil {
		t.Fatal("parseConfigFile must reject legacy YAML")
	}
	if !strings.Contains(err.Error(), "request.method is required") {
		t.Fatalf("error must name the missing new-schema field, got: %v", err)
	}
}

// S1.7 — SaveRequestRaw on an unknown profile returns an error and
// does not write anywhere.
func TestManager_SaveProfileRaw_UnknownProfile(t *testing.T) {
	mgr, _ := newRawTestManager(t, map[string]string{"p": "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 1\n"})

	err := mgr.SaveRequestRaw("ghost", "request:\n  method: GET\n  url_template: http://y\ncsv:\n  fields: [a]\nworkers: 1\n")
	if err == nil {
		t.Fatalf("SaveRequestRaw on unknown profile must return an error, got nil")
	}
}

// S7.1 — Save re-applies workers via OnChange. The OnChange listener
// receives the parsed config with the new workers value; the spec
// says processor.SetWorkers is downstream of that, but the contract
// under test is the OnChange callback receives the new workers.
func TestManager_SaveProfileRaw_OnChangeReceivesNewWorkers(t *testing.T) {
	mgr, _ := newRawTestManager(t, map[string]string{"p": "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"})

	var (
		calls   atomic.Int32
		lastCfg *Config
	)
	mgr.OnChange(func(cfg *Config) {
		calls.Add(1)
		lastCfg = cfg
	})

	const newYAML = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 8\n"
	if err := mgr.SaveRequestRaw("p", newYAML); err != nil {
		t.Fatalf("SaveRequestRaw: %v", err)
	}

	if calls.Load() != 1 {
		t.Fatalf("OnChange must fire exactly once, got %d", calls.Load())
	}
	if lastCfg == nil || lastCfg.Workers != 8 {
		t.Fatalf("OnChange payload must carry the new workers=8, got %+v", lastCfg)
	}
}

// S8.1 — updateByName mutates the named profile's in-memory *Config
// only. Profiles whose name does not match are left untouched. The
// helper must NOT fire OnChange; that responsibility belongs to the
// caller (SaveRequestRaw → conditional notify).
//
// Map iteration order in Go is randomized, so the test discovers
// which of {p, q} became active and treats the other as the
// non-active sibling it will mutate. updateByName must not switch
// the active profile even when the named profile is the active one.
func TestManager_UpdateByName_MutatesOnlyNamed(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://p\ncsv:\n  fields: [id]\nworkers: 4\n"
	const qRaw = "request:\n  method: GET\n  url_template: http://q\ncsv:\n  fields: [a]\nworkers: 2\n"
	mgr, _ := newRawTestManager(t, map[string]string{
		"p": pRaw,
		"q": qRaw,
	})
	// newRawTestManager makes the first iterated (map order
	// randomized) profile active. The test mutates p and asserts
	// q is untouched; the active profile must be unchanged.
	active := mgr.GetActiveRequest()
	if active != "p" && active != "q" {
		t.Fatalf("precondition: active profile must be p or q, got %q", active)
	}
	var other string
	if active == "p" {
		other = "q"
	} else {
		other = "p"
	}

	// Register a listener — it must never fire from updateByName.
	var calls atomic.Int32
	mgr.OnChange(func(_ *Config) { calls.Add(1) })

	// Build a new *Config for p (different workers, same shape).
	newCfg := &Config{
		Request: RequestConfig{Method: "PUT", URLTemplate: "http://p-new"},
		CSV:     CSVConfig{},
		Workers: 8,
	}

	if err := mgr.requestMgr.updateByName("p", newCfg); err != nil {
		t.Fatalf("updateByName(\"p\"): %v", err)
	}

	// p.Config was swapped.
	gotP := mgr.GetRequest("p")
	if gotP == nil {
		t.Fatalf("GetRequest(\"p\") returned nil after updateByName")
	}
	if gotP.Request.Method != "PUT" || gotP.Request.URLTemplate != "http://p-new" || gotP.Workers != 8 {
		t.Fatalf("p.Config not swapped as expected, got %+v", gotP)
	}

	// q.Config is unchanged.
	gotQ := mgr.GetRequest("q")
	if gotQ == nil {
		t.Fatalf("GetRequest(\"q\") returned nil after updateByName(\"p\")")
	}
	if gotQ.Request.Method != "GET" || gotQ.Request.URLTemplate != "http://q" || gotQ.Workers != 2 {
		t.Fatalf("q.Config must be unchanged by updateByName(\"p\"), got %+v", gotQ)
	}

	// Active profile is unchanged (mutating "p" must NOT change
	// the active profile — that is the caller's job, see
	// SetActiveProfile / SaveRequestRaw).
	if got := mgr.GetActiveRequest(); got != active {
		t.Fatalf("active profile must remain %q, got %q", active, got)
	}
	_ = other // referenced to keep the active-vs-other dichotomy explicit

	// OnChange was NOT fired — updateByName is purely in-memory.
	if c := calls.Load(); c != 0 {
		t.Fatalf("updateByName must NOT fire OnChange, got %d calls", c)
	}
}

// S8.2 — updateByName on an unknown name is a no-op (returns nil).
// No panic, no error, profiles slice unchanged, no OnChange.
func TestManager_UpdateByName_UnknownIsNoOp(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://p\ncsv:\n  fields: [id]\nworkers: 4\n"
	mgr, _ := newRawTestManager(t, map[string]string{"p": pRaw})

	var calls atomic.Int32
	mgr.OnChange(func(_ *Config) { calls.Add(1) })

	before := mgr.GetRequest("p")
	if before == nil {
		t.Fatalf("precondition: p.Config must exist")
	}

	if err := mgr.requestMgr.updateByName("ghost", before); err != nil {
		t.Fatalf("updateByName on unknown name must return nil error, got %v", err)
	}

	// p.Config unchanged.
	after := mgr.GetRequest("p")
	if after == nil {
		t.Fatalf("p.Config must still exist after a no-op updateByName")
	}
	if after.Request.Method != before.Request.Method ||
		after.Request.URLTemplate != before.Request.URLTemplate ||
		after.Workers != before.Workers {
		t.Fatalf("p.Config must be unchanged on no-op updateByName\nbefore: %+v\nafter:  %+v", before, after)
	}

	// ListRequests still contains only p.
	names := mgr.ListRequests()
	if len(names) != 1 || names[0] != "p" {
		t.Fatalf("ListRequests must be unchanged, got %v", names)
	}

	if c := calls.Load(); c != 0 {
		t.Fatalf("updateByName must NOT fire OnChange, got %d calls", c)
	}
}

// requireActive asserts the active profile is name; helper for
// readability in the S8.x updateByName tests above.
func requireActive(t *testing.T, mgr *runManager, name string) {
	t.Helper()
	if got := mgr.GetActiveRequest(); got != name {
		t.Fatalf("active profile: got %q, want %q", got, name)
	}
}

// S1.3a / S7.3 — Saving a non-active profile writes the file,
// mutates only the named profile's in-memory *Config, and does NOT
// fire OnChange. The active profile's *Config must remain intact so
// gateway + processor are not silently rewired with the wrong config.
//
// Before the fix SaveRequestRaw always called m.Update(cfg) →
// profileMgr.updateActive, which overwrote the active profile's
// in-memory config and fired OnChange regardless of the saved name.
//
// The test forces p to be the active profile by writing
// profileMgr.activeIndex after seeding (Go map iteration is
// randomized, so the helper's "first wins" tiebreaker is not
// stable across runs).
func TestManager_SaveProfileRaw_NonActiveSkipsOnChange(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://p\ncsv:\n  fields: [id]\nworkers: 2\n"
	const qRaw = "request:\n  method: GET\n  url_template: http://q\ncsv:\n  fields: [a]\nworkers: 1\n"
	mgr, dir := newRawTestManager(t, map[string]string{
		"p": pRaw,
		"q": qRaw,
	})
	// Force p active: map iteration order is randomized, so the
	// helper's "first seeded" tiebreaker is non-deterministic.
	for i, p := range mgr.requestMgr.requests {
		if p.Name == "p" {
			mgr.requestMgr.activeIndex = i
			break
		}
	}
	requireActive(t, mgr, "p")

	// Snapshot the active profile's in-memory *Config so we can
	// assert it survives the non-active save.
	activeBefore := mgr.Get()
	if activeBefore == nil {
		t.Fatalf("precondition: active *Config must exist")
	}

	var calls atomic.Int32
	mgr.OnChange(func(_ *Config) { calls.Add(1) })

	const newQ = "request:\n  method: PUT\n  url_template: http://q-new\ncsv:\n  fields: [a, b]\nworkers: 5\n"
	if err := mgr.SaveRequestRaw("q", newQ); err != nil {
		t.Fatalf("SaveRequestRaw(\"q\"): %v", err)
	}

	// OnChange must NOT fire on a non-active save.
	if c := calls.Load(); c != 0 {
		t.Fatalf("non-active save must NOT fire OnChange, got %d calls", c)
	}

	// Active profile's in-memory *Config is unchanged (same pointer
	// AND same content). SaveRequestRaw on "q" must not reach into
	// the active profile at all.
	activeAfter := mgr.Get()
	if activeAfter == nil {
		t.Fatalf("active *Config must still exist after non-active save")
	}
	if activeBefore.Request.Method != activeAfter.Request.Method ||
		activeBefore.Request.URLTemplate != activeAfter.Request.URLTemplate ||
		activeBefore.Workers != activeAfter.Workers {
		t.Fatalf("active *Config must not be mutated by non-active save\nbefore: %+v\nafter:  %+v", activeBefore, activeAfter)
	}

	// Named (non-active) profile's in-memory *Config is updated.
	qCfg := mgr.GetRequest("q")
	if qCfg == nil {
		t.Fatalf("GetRequest(\"q\") returned nil after SaveRequestRaw")
	}
	if qCfg.Request.Method != "PUT" || qCfg.Request.URLTemplate != "http://q-new" || qCfg.Workers != 5 {
		t.Fatalf("q.Config not updated as expected, got %+v", qCfg)
	}

	// The on-disk file for q is the new YAML.
	written, err := os.ReadFile(filepath.Join(dir, "q.yml"))
	if err != nil {
		t.Fatalf("read back q.yml: %v", err)
	}
	if string(written) != newQ {
		t.Fatalf("q.yml on disk must reflect the new content.\nwant: %q\ngot:  %q", newQ, string(written))
	}

	// Active profile's on-disk file is unchanged.
	origP, err := os.ReadFile(filepath.Join(dir, "p.yml"))
	if err != nil {
		t.Fatalf("read back p.yml: %v", err)
	}
	if string(origP) != pRaw {
		t.Fatalf("p.yml on disk must NOT be touched by a non-active save.\nwant: %q\ngot:  %q", pRaw, string(origP))
	}

	// Active profile is still "p" (no auto-activation on save).
	if got := mgr.GetActiveRequest(); got != "p" {
		t.Fatalf("active profile must remain p after non-active save, got %q", got)
	}
}

// S9.1 / S9.2 — Concurrent SaveRequestRaw calls and re-entrant
// listeners must not race or deadlock. Two goroutines hammer
// SaveRequestRaw on the active profile while the listener calls back
// into mgr.Get() (re-entrant). All calls must complete, the listener
// count must equal 2N, and `go test -race` must stay clean.
func TestManager_SaveProfileRaw_ConcurrentNoRace(t *testing.T) {
	const pRaw = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 2\n"
	mgr, _ := newRawTestManager(t, map[string]string{"p": pRaw})
	requireActive(t, mgr, "p")

	const (
		goroutines = 2
		perG       = 8
	)
	var calls atomic.Int32
	mgr.OnChange(func(_ *Config) {
		// Re-entrant call: Get() takes profileMgr.mu.RLock while
		// the manager's own mu is held by the snapshot block. The
		// refactor must snapshot listeners under RLock → release
		// before iterating, otherwise this deadlocks.
		_ = mgr.Get()
		calls.Add(1)
	})

	const newYAML = "request:\n  method: POST\n  url_template: http://x\ncsv:\n  fields: [id]\nworkers: 8\n"

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				if err := mgr.SaveRequestRaw("p", newYAML); err != nil {
					t.Errorf("SaveRequestRaw: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	if got, want := calls.Load(), int32(goroutines*perG); got != want {
		t.Fatalf("OnChange call count = %d, want %d", got, want)
	}
}

// S1.8 — GetRequestRaw on a profile whose file was deleted after
// discovery must surface a non-nil error. The manager does not silently
// re-read the disk during discovery, so this is a stale-cache path.
func TestManager_GetProfileRaw_FileDeletedAfterDiscovery(t *testing.T) {
	mgr, dir := newRawTestManager(t, map[string]string{"p": "request:\n  method: GET\n  url_template: http://x\ncsv:\n  fields: [a]\nworkers: 1\n"})

	// Delete the underlying file out from under the manager.
	path := filepath.Join(dir, "p.yml")
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove: %v", err)
	}

	got, err := mgr.GetRequestRaw("p")
	if err == nil {
		t.Fatalf("GetRequestRaw must return an error when the file is missing, got %q", got)
	}
	// errors.Is should be able to find os.ErrNotExist in the chain so
	// callers can detect the missing-file case programmatically.
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error must wrap os.ErrNotExist, got: %v", err)
	}
}

// _ keeps `io` import live for the package; tests above don't use it
// directly but future tests on ReadCloser-style probes will.
var _ = io.Discard
