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
	"go.uber.org/mock/gomock"
)

// newTestProcessor builds a Processor with gomock-backed HttpGateway and
// RequestLogger and returns all three. The helper registers NO
// expectations — tests own the call-count assertions because Exec/Add/
// WriteToFile semantics differ per test (some are .Times(n), some
// MinTimes(1), some AnyTimes()).
//
// workers is currently always 1 in the tests; the parameter is
// retained so future tests can exercise the worker-pool path
// without changing the helper signature.
//
//nolint:unparam
func newTestProcessor(t *testing.T, csvCfg config.CSVConfig, workers int) (
	*processorImpl,
	*mock_processor.MockHttpGateway,
	*mock_processor.MockRequestLogger,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	gatewayMock := mock_processor.NewMockHttpGateway(ctrl)
	loggerMock := mock_processor.NewMockRequestLogger(ctrl)
	p := NewProcessor(csvCfg, gatewayMock, loggerMock, workers)
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

		csvCfg := config.CSVConfig{Fields: []string{"header1", "header2"}, Separator: ","}
		p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg, 1)
		gatewayMock.EXPECT().
			Exec(gomock.Any(), gomock.Any()).
			Do(func(arg0, arg1 any) {
				wg.Done()
			}).Times(2)
		loggerMock.EXPECT().Add(gomock.Any()).MinTimes(1)
		loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(2)

		p.Do(ctx, tempFile.Name())
		wg.Wait()
	})

	t.Run("Should send a request with only selected fields from the csv line", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)
		csvData := "header1,header2\nvalue1,value2\n"
		tempFile := createCsvFile(t, csvData)
		defer os.Remove(tempFile.Name())

		csvCfg := config.CSVConfig{Fields: []string{"header1"}, Separator: ","}
		p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg, 1)
		gatewayMock.EXPECT().
			Exec(gomock.Any(), map[string]string{"header1": "value1"}).
			Do(func(arg0, arg1 any) {
				wg.Done()
			}).Times(1)
		loggerMock.EXPECT().Add(gomock.Any()).MinTimes(1)
		loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(1)

		p.Do(context.Background(), tempFile.Name())
		wg.Wait()
	})
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

	csvCfg := config.CSVConfig{Fields: []string{"header1"}, Separator: ","}
	p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg, 1)
	gatewayMock.EXPECT().
		Exec(gomock.Any(), map[string]string{"header2": "value2"}).
		Do(func(arg0, arg1 any) {
			wg.Done()
		}).Times(1)
	loggerMock.EXPECT().Add(gomock.Any()).AnyTimes()
	loggerMock.EXPECT().WriteToFile(gomock.Any()).AnyTimes()

	// Swap the field filter before running Do. The next run must honor it.
	p.UpdateConfig(config.CSVConfig{Fields: []string{"header2"}, Separator: ","})

	p.Do(context.Background(), tempFile.Name())
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

	csvCfg := config.CSVConfig{Fields: []string{"header1"}, Separator: ","}
	p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg, 1)

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

	p.Do(context.Background(), tempFile.Name())

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
			assert.Equal(t, 200, m.StatusCode)
			assert.Equal(t, "GET", m.Method)
			assert.Equal(t, "https://example.com/users/1", m.URL)
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

	csvCfg := config.CSVConfig{Fields: []string{"header1"}, Separator: ","}
	p, gatewayMock, loggerMock := newTestProcessor(t, csvCfg, 1)

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
		if m.Type == logs.LogTypeClientError {
			select {
			case <-execDone:
				// already closed
			default:
				close(execDone)
			}
		}
	}).AnyTimes()
	loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(1)

	p.Do(context.Background(), tempFile.Name())

	select {
	case <-execDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the 4xx error log message")
	}

	mu.Lock()
	defer mu.Unlock()
	var sawClientError bool
	for _, m := range added {
		if m.Type == logs.LogTypeClientError {
			assert.Equal(t, 404, m.StatusCode)
			assert.Equal(t, "GET", m.Method)
			assert.Equal(t, "https://example.com/missing", m.URL)
			assert.Equal(t, []byte(`{"error":"not found"}`), m.Body)
			sawClientError = true
		}
	}
	assert.True(t, sawClientError, "worker must call logger.Add with LogTypeClientError for a 4xx response")
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
