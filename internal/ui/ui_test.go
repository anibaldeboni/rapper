package ui_test

import (
	"testing"

	manager_mock "github.com/anibaldeboni/rapper/internal/config/mock"
	mock_log "github.com/anibaldeboni/rapper/internal/logs/mock"
	mock_processor "github.com/anibaldeboni/rapper/internal/processor/mock"
	"github.com/anibaldeboni/rapper/internal/ui"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestNewUI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logManagerMock := mock_log.NewMockLogger(ctrl)
	processorMock := mock_processor.NewMockProcessor(ctrl)
	configMgrMock := manager_mock.NewMockManager(ctrl)

	t.Run("When the path contains CSV files", func(t *testing.T) {
		filePaths := []string{"../../tests/example.csv"}

		// Mock Logger.Get() which is called during UI initialization in updateLogs()
		logManagerMock.EXPECT().Get().Return([]string{}).AnyTimes()
		// Mock ConfigManager.Get() which is called during settings view initialization
		configMgrMock.EXPECT().Get().Return(nil).AnyTimes()
		// Mock Processor.GetWorkerCount() which is called during workers view initialization
		processorMock.EXPECT().GetWorkerCount().Return(1).AnyTimes()

		c := ui.NewApp(filePaths, processorMock, logManagerMock, configMgrMock)

		assert.NotNil(t, c)
	})
}
