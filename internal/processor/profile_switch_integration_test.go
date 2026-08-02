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
//  1. Two YAML profiles in a temp dir with different request configurations.
//  2. Manager.OnChange subscribed to BOTH the HTTP gateway (UpdateConfig) and the
//     CSV processor (UpdateConfig).
//  3. SetActiveProfile("profile2") → processor must use the active CSV
//     configuration on the next Do.
//
// Before the fix this test fails because SetActiveProfile never fires OnChange
// and the processor is never subscribed.
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
workers: 2
`)
	profile2 := []byte(`request:
  method: POST
  url_template: https://api2.example
  body_template: "{}"
  headers:
    Content-Type: application/json
csv:
  separator: ","
workers: 8
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
	gatewayMock.EXPECT().TemplateFields().Return([]string{}).AnyTimes()
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

	mgr, err := config.NewRunManager(dir)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	startCfg := mgr.Get()
	if startCfg == nil {
		t.Fatalf("expected active config on manager init")
	}

	csvProcessor := NewProcessor(startCfg.CSV, gatewayMock, loggerMock)

	// Wire the same callback main.go uses.
	mgr.OnChange(func(newCfg *config.Config) {
		_ = gatewayMock.UpdateConfig(
			newCfg.Request.Method,
			newCfg.Request.URLTemplate,
			newCfg.Request.BodyTemplate,
			newCfg.Request.Headers,
		)
		csvProcessor.UpdateConfig(newCfg.CSV)
	})

	// Switch to profile2.
	if err := mgr.SetActiveRequest("profile2"); err != nil {
		t.Fatalf("SetActiveRequest: %v", err)
	}
	activeCfg := mgr.Get()

	// Write a CSV and run it. All columns remain available to the gateway.
	csvPath := filepath.Join(dir, "input.csv")
	if err := os.WriteFile(csvPath, []byte("col_a,col_b\nval_a,val_b\n"), 0o600); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newCtx, newCancel, err := csvProcessor.Do(ctx, csvPath, activeCfg.Workers)
	if err != nil {
		t.Fatalf("processor.Do: %v", err)
	}
	if newCtx == nil {
		t.Fatalf("processor.Do returned nil context (csv read failed?)")
	}
	defer newCancel()
	if got := csvProcessor.GetMetrics().ActiveWorkers; got != 8 {
		t.Fatalf("expected active workers from switched profile to be 8, got %d", got)
	}

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
	if _, ok := got["col_a"]; !ok {
		t.Fatalf("expected col_a in gateway data map after profile2 switch, got %v", got)
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
