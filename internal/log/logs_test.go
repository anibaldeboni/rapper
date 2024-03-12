package log_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/stretchr/testify/assert"
)

type logMessage struct {
	message string
}

func (l logMessage) String() string {
	return l.message
}

func TestGet(t *testing.T) {
	t.Run("Should return all logs", func(t *testing.T) {
		logManager := log.NewLogManager()
		var expectedLogs []string
		logs := []logMessage{{message: "log 1"}, {message: "log 2"}, {message: "log 3"}}
		for _, log := range logs {
			logManager.Add(log)
			expectedLogs = append(expectedLogs, log.String())
		}

		got := logManager.Get()

		assert.Equal(t, expectedLogs, got)
	})

	t.Run("Should return an empty slice when there are no logs", func(t *testing.T) {
		logManager := log.NewLogManager()
		logs := logManager.Get()

		assert.Empty(t, logs)
	})
}

func TestHasNewLogs(t *testing.T) {
	t.Run("Should return true when there are new logs", func(t *testing.T) {
		logManager := log.NewLogManager()
		logs := []logMessage{{message: "log 1"}, {message: "log 2"}, {message: "log 3"}}
		for _, log := range logs {
			logManager.Add(log)
		}

		first := logManager.HasNewLogs()
		second := logManager.HasNewLogs()

		assert.True(t, first)
		assert.False(t, second)
	})
}
