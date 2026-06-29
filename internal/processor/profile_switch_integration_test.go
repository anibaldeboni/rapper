package processor

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/anibaldeboni/rapper/internal/config"
	mock_processor "github.com/anibaldeboni/rapper/internal/processor/mock"
	"go.uber.org/mock/gomock"
)

// TestProfileSwitch_EndToEnd proves the profile-switch hot-reload wiring:
//  1. Two YAML profiles in a temp dir with different CSV fields.
//  2. Manager.OnChange subscribed to BOTH the HTTP gateway (UpdateConfig) and the
//     CSV processor (UpdateConfig + SetWorkers).
//  3. SetActiveProfile("profile2") → processor must filter rows using profile2's
//     CSV fields on the next Do.
//
// Before the fix this test fails because SetActiveProfile never fires OnChange
// and the processor is never subscribed. The processor keeps profile1's
// field filter and emits `{"col_a": "val_a"}` instead of `{"col_b": "val_b"}`.
func TestProfileSwitch_EndToEnd(t *testing.T) {
	dir := t.TempDir()

	profile1 := []byte(`request:
  method: POST
  url_template: https://api1.example
  body_template: "{}"
  headers:
    Content-Type: application/json
csv:
  separator: ","
  fields: [col_a]
workers: 1
`)
	profile2 := []byte(`request:
  method: POST
  url_template: https://api2.example
  body_template: "{}"
  headers:
    Content-Type: application/json
csv:
  separator: ","
  fields: [col_b]
workers: 1
`)
	if err := os.WriteFile(filepath.Join(dir, "profile1.yml"), profile1, 0o600); err != nil {
		t.Fatalf("write profile1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "profile2.yml"), profile2, 0o600); err != nil {
		t.Fatalf("write profile2: %v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gatewayMock := mock_processor.NewMockHttpGateway(ctrl)
	// Capture the data map the processor hands to the gateway.
	var (
		capturedMu sync.Mutex
		captured   map[string]string
		wg         sync.WaitGroup
	)
	wg.Add(1)
	gatewayMock.EXPECT().
		UpdateConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	gatewayMock.EXPECT().
		Exec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, data map[string]string) (struct{}, error) {
			capturedMu.Lock()
			captured = data
			capturedMu.Unlock()
			wg.Done()
			return struct{}{}, nil
		}).
		Times(1)

	loggerMock := mock_processor.NewMockRequestLogger(ctrl)
	loggerMock.EXPECT().Add(gomock.Any()).AnyTimes()
	loggerMock.EXPECT().WriteToFile(gomock.Any()).AnyTimes()

	mgr, err := config.NewManager(dir)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	startCfg := mgr.Get()
	if startCfg == nil {
		t.Fatalf("expected active config on manager init")
	}

	csvProcessor := NewProcessor(startCfg.CSV, gatewayMock, loggerMock, 1)

	// Wire the same callback main.go uses.
	mgr.OnChange(func(newCfg *config.Config) {
		_ = gatewayMock.UpdateConfig(
			newCfg.Request.Method,
			newCfg.Request.URLTemplate,
			newCfg.Request.BodyTemplate,
			newCfg.Request.Headers,
		)
		csvProcessor.UpdateConfig(newCfg.CSV)
		if newCfg.Workers > 0 {
			csvProcessor.SetWorkers(newCfg.Workers)
		}
	})

	// Switch to profile2.
	if err := mgr.SetActiveProfile("profile2"); err != nil {
		t.Fatalf("SetActiveProfile: %v", err)
	}

	// Write a CSV and run it. Both fields are present in the row; the active
	// profile's field filter decides which one(s) reach the gateway.
	csvPath := filepath.Join(dir, "input.csv")
	if err := os.WriteFile(csvPath, []byte("col_a,col_b\nval_a,val_b\n"), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newCtx, newCancel := csvProcessor.Do(ctx, csvPath)
	if newCtx == nil {
		t.Fatalf("processor.Do returned nil context (csv read failed?)")
	}
	defer newCancel()

	if !waitFor(2*time.Second, func() bool {
		capturedMu.Lock()
		defer capturedMu.Unlock()
		return captured != nil
	}) {
		t.Fatalf("timed out waiting for gateway.Exec to be called")
	}

	capturedMu.Lock()
	got := captured
	capturedMu.Unlock()
	if _, ok := got["col_b"]; !ok {
		t.Fatalf("expected col_b in gateway data map after profile2 switch, got %v", got)
	}
	if _, ok := got["col_a"]; ok {
		t.Fatalf("did not expect col_a in gateway data map after profile2 switch (filter not applied), got %v", got)
	}
}

func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
