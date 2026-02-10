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

func TestProcessor_Do(t *testing.T) {
	t.Run("Should send a request for each csv line", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := context.Background()
		wg := sync.WaitGroup{}
		wg.Add(2)
		gatewayMock := mock_processor.NewMockHttpGateway(ctrl)
		gatewayMock.EXPECT().
			Exec(gomock.Any(), gomock.Any()).
			Do(func(arg0, arg1 any) {
				wg.Done()
			}).Times(2)
		loggerMock := mock_processor.NewMockRequestLogger(ctrl)
		loggerMock.EXPECT().Add(gomock.Any()).MinTimes(1)
		loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(2)

		csvData := "header1,header2\nvalue1,value2\nvalue3,value4\n"
		tempFile := createCsvFile(t, csvData)
		defer os.Remove(tempFile.Name())

		p := NewProcessor(config.CSVConfig{Fields: []string{"header1", "header2"}, Separator: ","}, gatewayMock, loggerMock, 1)

		p.Do(ctx, tempFile.Name())
		wg.Wait()
	})

	t.Run("Should send a request with only selected fields from the csv line", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		wg := sync.WaitGroup{}
		wg.Add(1)
		gatewayMock := mock_processor.NewMockHttpGateway(ctrl)
		gatewayMock.EXPECT().
			Exec(gomock.Any(), map[string]string{"header1": "value1"}).
			Do(func(arg0, arg1 any) {
				wg.Done()
			}).Times(1)
		loggerMock := mock_processor.NewMockRequestLogger(ctrl)
		loggerMock.EXPECT().Add(gomock.Any()).MinTimes(1)
		loggerMock.EXPECT().WriteToFile(gomock.Any()).Times(1)

		csvData := "header1,header2\nvalue1,value2\n"
		tempFile := createCsvFile(t, csvData)
		defer os.Remove(tempFile.Name())

		p := NewProcessor(config.CSVConfig{Fields: []string{"header1"}, Separator: ","}, gatewayMock, loggerMock, 1)

		p.Do(context.Background(), tempFile.Name())
		wg.Wait()
	})
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
