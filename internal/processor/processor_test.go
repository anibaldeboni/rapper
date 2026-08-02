package processor

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	mock_processor "github.com/anibaldeboni/rapper/internal/processor/mock"
	"github.com/anibaldeboni/rapper/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// newTestProcessor builds a Processor with gomock-backed HttpGateway and
// RequestLogger and returns all three. The helper registers NO
// expectations — tests own the call-count assertions because Exec/Add/
// WriteToFile semantics differ per test (some are .Times(n), some
// MinTimes(1), some AnyTimes()).
func newTestProcessor(t *testing.T, csvCfg config.CSVConfig) (
	*processorImpl,
	*mock_processor.MockHttpGateway,
	*mock_processor.MockRequestLogger,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	gatewayMock := mock_processor.NewMockHttpGateway(ctrl)
	loggerMock := mock_processor.NewMockRequestLogger(ctrl)
	p := NewProcessor(csvCfg, gatewayMock, loggerMock)
	return p, gatewayMock, loggerMock
}

func TestProcessor_Do(t *testing.T) {
	t.Run("Should send a request for each csv line", func(t *testing.T) {
		ctx := context.Background()
		wg := sync.WaitGroup{}
		wg.Add(2)
		csvData := "header1,header2\nvalue1,value2\nvalue3,value4\n"
		tempFile := createCsvFile(t, csvData)
		defer os.Remove(tempFile.Name())

		csvCfg := config.CSVConfig{Separator: ","}
		p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg)
		gatewayMock.EXPECT().TemplateFields().Return([]string{"header1", "header2"})
		gatewayMock.EXPECT().
			Exec(gomock.Any(), gomock.Any()).
			Do(func(arg0, arg1 any) {
				wg.Done()
			}).Times(2)
		loggerMock.EXPECT().Add(gomock.Any()).MinTimes(1)
		loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(2)

		_, _, err := p.Do(ctx, tempFile.Name(), 1)
		require.NoError(t, err)
		wg.Wait()
	})

	t.Run("Should send a request with all fields from the csv line", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)
		csvData := "header1,header2\nvalue1,value2\n"
		tempFile := createCsvFile(t, csvData)
		defer os.Remove(tempFile.Name())

		csvCfg := config.CSVConfig{Separator: ","}
		p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg)
		gatewayMock.EXPECT().TemplateFields().Return([]string{"header1"})
		gatewayMock.EXPECT().
			Exec(gomock.Any(), map[string]string{"header1": "value1", "header2": "value2"}).
			Do(func(arg0, arg1 any) {
				wg.Done()
			}).Times(1)
		loggerMock.EXPECT().Add(gomock.Any()).MinTimes(1)
		loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(1)

		_, _, err := p.Do(context.Background(), tempFile.Name(), 1)
		require.NoError(t, err)
		wg.Wait()
	})
}

func TestProcessor_Do_PreflightRejectsMissingFieldsBeforeExec(t *testing.T) {
	tempFile := createCsvFile(t, "id\n1\n")

	p, gatewayMock, loggerMock := newTestProcessor(t, config.CSVConfig{Separator: ","})
	gatewayMock.EXPECT().TemplateFields().Return([]string{"id", "name"})
	gatewayMock.EXPECT().Exec(gomock.Any(), gomock.Any()).Times(0)
	loggerMock.EXPECT().Add(gomock.Any()).AnyTimes()

	ctx, cancel, err := p.Do(context.Background(), tempFile.Name(), 1)
	require.Nil(t, ctx)
	require.Nil(t, cancel)
	var missing *MissingFieldsError
	require.ErrorAs(t, err, &missing)
	assert.Equal(t, []string{"name"}, missing.Missing)
}

// TestProcessor_UpdateConfig proves that UpdateConfig atomically replaces the
// processor's CSV configuration. Before the fix the method does not exist and
// the test fails to compile.
func TestProcessor_UpdateConfig(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	csvData := "header1,header2\nvalue1,value2\n"
	tempFile := createCsvFile(t, csvData)
	defer os.Remove(tempFile.Name())

	csvCfg := config.CSVConfig{Separator: ","}
	p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg)
	gatewayMock.EXPECT().TemplateFields().Return([]string{"header2"})
	gatewayMock.EXPECT().
		Exec(gomock.Any(), map[string]string{"header1": "value1", "header2": "value2"}).
		Do(func(arg0, arg1 any) {
			wg.Done()
		}).Times(1)
	loggerMock.EXPECT().Add(gomock.Any()).AnyTimes()
	loggerMock.EXPECT().WriteToFile(gomock.Any()).AnyTimes()

	// Swap the separator before running Do. The next run must honor it.
	p.UpdateConfig(config.CSVConfig{Separator: ","})

	_, _, err := p.Do(context.Background(), tempFile.Name(), 1)
	require.NoError(t, err)
	wg.Wait()
}

// TestProcessor_Do_LogsSuccessOnTwoXX proves the success path: when
// the gateway returns a 2xx response, the worker calls
// logger.Add(logs.NewHTTPMessage(res)). Before the fix the success
// branch was missing entirely, so successful runs were invisible in
// the TUI logs view.
func TestProcessor_Do_LogsSuccessOnTwoXX(t *testing.T) {
	execDone := make(chan struct{})
	csvData := "header1\nvalue1\n"
	tempFile := createCsvFile(t, csvData)
	defer os.Remove(tempFile.Name())

	csvCfg := config.CSVConfig{Separator: ","}
	p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg)
	gatewayMock.EXPECT().TemplateFields().Return([]string{"header1"})

	gatewayMock.EXPECT().
		Exec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ map[string]string) (web.Response, error) {
			return web.Response{
				Method:     "GET",
				URL:        "https://example.com/users/1",
				StatusCode: 200,
				Body:       []byte(`{"id":1}`),
			}, nil
		}).Times(1)

	var (
		mu    sync.Mutex
		added []logs.LogMessage
	)
	loggerMock.EXPECT().Add(gomock.Any()).DoAndReturn(func(m logs.LogMessage) {
		mu.Lock()
		added = append(added, m)
		mu.Unlock()
		if m.Type == logs.LogTypeSuccess {
			select {
			case <-execDone:
				// already closed
			default:
				close(execDone)
			}
		}
	}).AnyTimes()
	loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(1)

	_, _, err := p.Do(context.Background(), tempFile.Name(), 1)
	require.NoError(t, err)

	select {
	case <-execDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the success-path log message")
	}

	mu.Lock()
	defer mu.Unlock()
	var sawSuccess bool
	for _, m := range added {
		if m.Type == logs.LogTypeSuccess {
			// NOTE: source behavior as of 2026-07-10 — see decision #178.
			// BadgeIcon is built via strconv.Itoa (string) and Text is
			// "METHOD URL".
			assert.Equal(t, "200", m.BadgeIcon)
			assert.Equal(t, "GET https://example.com/users/1", m.Text)
			sawSuccess = true
		}
	}
	assert.True(t, sawSuccess, "worker must call logger.Add with LogTypeSuccess for a 2xx response")
}

// TestProcessor_Do_LogsClientErrorOnFourXX proves that the worker
// surfaces 4xx HTTP responses as logs.NewHTTPMessage — the TUI
// renderer relies on LogTypeClientError to color the row and on
// the populated Body to show the response on Enter. Before the
// fix the default branch called NewGeneralMessage (no body, no
// LogType), so 4xx errors had no detail view.
func TestProcessor_Do_LogsClientErrorOnFourXX(t *testing.T) {
	execDone := make(chan struct{})
	csvData := "header1\nvalue1\n"
	tempFile := createCsvFile(t, csvData)
	defer os.Remove(tempFile.Name())

	csvCfg := config.CSVConfig{Separator: ","}
	p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg)
	gatewayMock.EXPECT().TemplateFields().Return([]string{"header1"})

	gatewayMock.EXPECT().
		Exec(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ map[string]string) (web.Response, error) {
			return web.Response{
				Method:     "GET",
				URL:        "https://example.com/missing",
				StatusCode: 404,
				Body:       []byte(`{"error":"not found"}`),
			}, nil
		}).Times(1)

	var (
		mu    sync.Mutex
		added []logs.LogMessage
	)
	loggerMock.EXPECT().Add(gomock.Any()).DoAndReturn(func(m logs.LogMessage) {
		mu.Lock()
		added = append(added, m)
		mu.Unlock()
		if m.Type == logs.LogTypeWarning {
			select {
			case <-execDone:
				// already closed
			default:
				close(execDone)
			}
		}
	}).AnyTimes()
	loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(1)

	_, _, err := p.Do(context.Background(), tempFile.Name(), 1)
	require.NoError(t, err)

	select {
	case <-execDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the 4xx error log message")
	}

	mu.Lock()
	defer mu.Unlock()
	var sawClientError bool
	for _, m := range added {
		if m.Type == logs.LogTypeWarning {
			// NOTE: source behavior as of 2026-07-10 — see decision #178.
			// BadgeIcon is a string, Text is "GET URL" (mock uses GET),
			// and Details is a pretty-printed ANSI-colored JSON string.
			assert.Equal(t, "404", m.BadgeIcon)
			assert.Equal(t, "GET https://example.com/missing", m.Text)
			assert.Contains(t, m.Details, "not found",
				"Details must contain the original body content (pretty.Color wraps it in ANSI)")
			sawClientError = true
		}
	}
	assert.True(t, sawClientError, "worker must call logger.Add with LogTypeClientError for a 4xx response")
}

func TestProcessor_Do_ResolvesWorkerCountPerRun(t *testing.T) {
	configuredWorkers := 4
	if configuredWorkers > maxWorkers {
		configuredWorkers = maxWorkers
	}

	tests := []struct {
		name    string
		workers int
		want    int
	}{
		{name: "zero falls back to one", workers: 0, want: 1},
		{name: "negative falls back to one", workers: -3, want: 1},
		{name: "positive value is retained", workers: 4, want: configuredWorkers},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := createCsvFile(t, "header1\n")
			defer os.Remove(tempFile.Name())

			csvCfg := config.CSVConfig{Separator: ","}
			p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg)
			gatewayMock.EXPECT().TemplateFields().Return([]string{"header1"})
			loggerMock.EXPECT().Add(gomock.Any()).AnyTimes()

			ctx, cancel, err := p.Do(context.Background(), tempFile.Name(), tt.workers)
			require.NoError(t, err)
			require.NotNil(t, ctx)
			defer cancel()

			assert.Equal(t, tt.want, p.GetMetrics().ActiveWorkers)
		})
	}
}

func createCsvFile(t *testing.T, csvData string) *os.File {
	tempFile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	_, err = tempFile.WriteString(csvData)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	return tempFile
}
