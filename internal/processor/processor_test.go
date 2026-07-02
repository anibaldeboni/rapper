package processor

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/anibaldeboni/rapper/internal/config"
	mock_processor "github.com/anibaldeboni/rapper/internal/processor/mock"
	"go.uber.org/mock/gomock"
)

// newTestProcessor builds a Processor with gomock-backed HttpGateway and
// RequestLogger and returns all three. The helper registers NO
// expectations — tests own the call-count assertions because Exec/Add/
// WriteToFile semantics differ per test (some are .Times(n), some
// MinTimes(1), some AnyTimes()).
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
