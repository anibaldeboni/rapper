package logs_test

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Run("Should return all logs", func(t *testing.T) {
		logger := logs.NewLogger("")
		msgs := []logs.LogMessage{
			logs.NewGeneralMessage("", "", "log 1"),
			logs.NewGeneralMessage("", "", "log 2"),
			logs.NewGeneralMessage("", "", "log 3"),
		}
		for _, m := range msgs {
			logger.Add(m)
		}

		got := logger.Get()

		assert.Equal(t, msgs, got, "Get must return the messages in insertion order")
	})

	t.Run("Should return an empty slice when there are no logs", func(t *testing.T) {
		logger := logs.NewLogger("")

		got := logger.Get()

		assert.Empty(t, got)
	})
}

// TestClear proves Clear empties the buffer and a subsequent Get
// returns an empty slice. After Clear, fresh Add calls land in the
// buffer as if the logger were just created — Clear is the only
// supported way to reset state mid-run.
func TestClear(t *testing.T) {
	logger := logs.NewLogger("")
	logger.Add(logs.NewGeneralMessage("", "", "first"))
	logger.Add(logs.NewGeneralMessage("", "", "second"))
	require.Len(t, logger.Get(), 2, "precondition: two messages were added")

	logger.Clear()

	assert.Empty(t, logger.Get(), "Get must be empty after Clear")

	// A subsequent Add starts a new run cleanly.
	logger.Add(logs.NewGeneralMessage("", "", "third"))
	assert.Len(t, logger.Get(), 1)
}

// TestGet_ReturnsCopy proves Get hands out a snapshot, not a live
// reference into the internal buffer. Mutating the returned slice
// must not affect the logger's state.
func TestGet_ReturnsCopy(t *testing.T) {
	logger := logs.NewLogger("")
	logger.Add(logs.NewGeneralMessage("", "", "keep me"))

	got := logger.Get()
	got[0] = logs.NewGeneralMessage("", "", "overwritten")

	assert.Equal(t, "keep me", logger.Get()[0].Text,
		"mutating the Get() result must not affect the internal buffer")
}

// TestLogger_Add_ConcurrentSafe proves Add is safe under concurrent
// writers. With the mutex in place every message lands in the buffer
// exactly once; without it, lost updates and even a data race on
// the slice header are possible.
func TestLogger_Add_ConcurrentSafe(t *testing.T) {
	const writers = 8
	const perWriter = 100

	logger := logs.NewLogger("")

	var wg sync.WaitGroup
	wg.Add(writers)
	for w := range writers {
		go func(id int) {
			defer wg.Done()
			for i := range perWriter {
				logger.Add(logs.NewGeneralMessage(
					"", "Concurrent",
					"writer-"+string(rune('A'+id))+"-"+string(rune('0'+i%10)),
				))
			}
		}(w)
	}
	wg.Wait()

	assert.Len(t, logger.Get(), writers*perWriter,
		"every Add must land in the buffer; lost updates indicate a race")
}

type testLog struct {
	Message string `json:"message"`
}

func (t testLog) Bytes() []byte {
	return []byte("{\"message\":\"" + t.Message + "\"}")
}

func TestWriteToFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_output_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	stream := logs.NewLogger(tmpFile.Name())

	line := testLog{Message: "test message"}
	stream.WriteToFile(line)

	// Read the contents of the temporary file
	file, err := os.Open(tmpFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	// Read the log message from the file
	var loggedMessage testLog
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&loggedMessage)
	assert.NoError(t, err)

	// Assert that the logged message matches the sent log message
	assert.Equal(t, &line, &loggedMessage)
}
