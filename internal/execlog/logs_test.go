package execlog_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/internal/execlog"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Run("Should return all logs", func(t *testing.T) {
		logManager := execlog.NewLogManager()
		var expectedLogs []string
		logs := []execlog.Message{
			execlog.NewMessage().WithMessage("log 1"),
			execlog.NewMessage().WithMessage("log 2"),
			execlog.NewMessage().WithMessage("log 3"),
		}
		for _, log := range logs {
			logManager.Add(log)
			expectedLogs = append(expectedLogs, log.String())
		}

		got := logManager.Get()

		assert.Equal(t, expectedLogs, got)
	})

	t.Run("Should return an empty slice when there are no logs", func(t *testing.T) {
		logManager := execlog.NewLogManager()
		logs := logManager.Get()

		assert.Empty(t, logs)
	})
}

func TestHasNewLogs(t *testing.T) {
	t.Run("Should return true when there are new logs", func(t *testing.T) {
		logManager := execlog.NewLogManager()
		logs := []execlog.Message{
			execlog.NewMessage().WithMessage("log 1"),
			execlog.NewMessage().WithMessage("log 2"),
			execlog.NewMessage().WithMessage("log 3"),
		}
		for _, log := range logs {
			logManager.Add(log)
		}

		first := logManager.HasNewLogs()
		second := logManager.HasNewLogs()

		assert.True(t, first)
		assert.False(t, second)
	})
}
