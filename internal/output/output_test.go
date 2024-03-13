package output_test

import (
	"encoding/json"
	"os"
	"testing"

	mock_log "github.com/anibaldeboni/rapper/internal/log/mock"
	"github.com/anibaldeboni/rapper/internal/output"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListen(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_output_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	ctrl := gomock.NewController(t)
	logManagerMock := mock_log.NewMockManager(ctrl)
	logManagerMock.EXPECT().Add(gomock.Any()).Times(0)

	stream := output.New(tmpFile.Name(), logManagerMock)

	line := output.Line{Status: 200, URL: "http://example.com"}
	stream.Send(line)

	// Read the contents of the temporary file
	file, err := os.Open(tmpFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	// Read the log message from the file
	var loggedMessage output.Line
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&loggedMessage)
	assert.NoError(t, err)

	// Assert that the logged message matches the sent log message
	assert.Equal(t, &line, &loggedMessage)
}
