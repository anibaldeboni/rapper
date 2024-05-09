package logs_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Run("Should return all logs", func(t *testing.T) {
		logger := logs.NewLoggger("")
		var expectedLogs []string
		logs := []logs.Message{
			logs.NewMessage().WithMessage("log 1"),
			logs.NewMessage().WithMessage("log 2"),
			logs.NewMessage().WithMessage("log 3"),
		}
		for _, log := range logs {
			logger.Add(log)
			expectedLogs = append(expectedLogs, log.String())
		}

		got := logger.Get()

		assert.Equal(t, expectedLogs, got)
	})

	t.Run("Should return an empty slice when there are no logs", func(t *testing.T) {
		logger := logs.NewLoggger("")
		logs := logger.Get()

		assert.Empty(t, logs)
	})
}

func TestHasNewLogs(t *testing.T) {
	t.Run("Should return true when there are new logs", func(t *testing.T) {
		logger := logs.NewLoggger("")
		logs := []logs.Message{
			logs.NewMessage().WithMessage("log 1"),
			logs.NewMessage().WithMessage("log 2"),
			logs.NewMessage().WithMessage("log 3"),
		}
		for _, log := range logs {
			logger.Add(log)
		}

		first := logger.HasNewLogs()
		second := logger.HasNewLogs()

		assert.True(t, first)
		assert.False(t, second)
	})
}

func TestWriteToFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_output_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	stream := logs.NewLoggger(tmpFile.Name())

	line := logs.Line{Status: 200, URL: "http://example.com"}
	stream.Write(line)

	// Read the contents of the temporary file
	file, err := os.Open(tmpFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	// Read the log message from the file
	var loggedMessage logs.Line
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&loggedMessage)
	assert.NoError(t, err)

	// Assert that the logged message matches the sent log message
	assert.Equal(t, &line, &loggedMessage)
}
