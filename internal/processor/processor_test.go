package processor

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/anibaldeboni/rapper/internal/config"
	mock_log "github.com/anibaldeboni/rapper/internal/log/mock"
	mock_web "github.com/anibaldeboni/rapper/internal/web/mock"
	"go.uber.org/mock/gomock"
)

func TestProcessor_mapCSV(t *testing.T) {
	t.Run("Should send a request for each csv line", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		wg := sync.WaitGroup{}
		wg.Add(2)
		gatewayMock := mock_web.NewMockHttpGateway(ctrl)
		gatewayMock.EXPECT().
			Exec(gomock.Any()).
			Do(func(arg0 interface{}) {
				wg.Done()
			}).Times(2)
		logsManagerMock := mock_log.NewMockLogManager(ctrl)
		logsManagerMock.EXPECT().Add(gomock.Any()).MinTimes(1)

		csvData := "header1,header2\nvalue1,value2\nvalue3,value4\n"
		tempFile := createCsvFile(t, csvData)
		defer os.Remove(tempFile.Name())

		ctx, cancel := context.WithCancel(context.Background())
		p := New(config.CSV{Fields: []string{"header1", "header2"}, Separator: ","}, gatewayMock, "", logsManagerMock)

		p.Do(ctx, cancel, tempFile.Name())
		wg.Wait()
	})

	t.Run("Should send a request with only selected fields from the csv line", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		wg := sync.WaitGroup{}
		wg.Add(1)
		gatewayMock := mock_web.NewMockHttpGateway(ctrl)
		gatewayMock.EXPECT().
			Exec(map[string]string{"header1": "value1"}).
			Do(func(arg0 any) {
				wg.Done()
			}).Times(1)
		logsManagerMock := mock_log.NewMockLogManager(ctrl)
		logsManagerMock.EXPECT().Add(gomock.Any()).MinTimes(1)

		csvData := "header1,header2\nvalue1,value2\n"
		tempFile := createCsvFile(t, csvData)
		defer os.Remove(tempFile.Name())

		ctx, cancel := context.WithCancel(context.Background())
		p := New(config.CSV{Fields: []string{"header1"}, Separator: ","}, gatewayMock, "", logsManagerMock)

		p.Do(ctx, cancel, tempFile.Name())
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
