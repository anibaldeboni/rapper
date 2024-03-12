package output

import (
	"encoding/json"
	"os"
	"testing"

	mock_log "github.com/anibaldeboni/rapper/internal/log/mock"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_output_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a channel to simulate log messages
	logs := make(chan Message)

	logManagerMock := mock_log.NewMockLogManager(nil)

	// Create an instance of the streamIpml struct
	o := &streamImpl{
		filePath: tmpFile.Name(),
		ch:       logs,
		logs:     logManagerMock,
	}

	// Start listening for log messages
	go listen(o)

	// Send a log message to the channel
	log := Message{Status: 200, URL: "http://example.com"}
	logs <- log

	// Close the channel to stop listening
	close(logs)

	// Read the contents of the temporary file
	file, err := os.Open(tmpFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	// Read the log message from the file
	var loggedMessage Message
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&loggedMessage)
	assert.NoError(t, err)

	// Assert that the logged message matches the sent log message
	assert.Equal(t, &log, &loggedMessage)
}
