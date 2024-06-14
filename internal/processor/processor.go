package processor

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/anibaldeboni/rapper/internal/web"
)

var (
	_           Processor = (*processorImpl)(nil)
	reqCount    atomic.Uint64
	errCount    atomic.Uint64
	linesCount  atomic.Uint64
	MAX_WORKERS = runtime.NumCPU()
)

//go:generate mockgen -destination mock/processor_mock.go github.com/anibaldeboni/rapper/internal/processor Processor
type Processor interface {
	Do(ctx context.Context, cancel func(), filePath string)
}

type csvLineMap map[string]string

type processorImpl struct {
	csvConfig config.CSV
	gateway   web.HttpGateway
	logger    logs.Logger
	workers   int
}

// NewProcessor creates a new instance of the Processor interface.
// It takes in the following parameters:
// - cfg: The CSV configuration.
// - hg: The HTTP gateway.
// - logger: The logger.
// - workers: The number of workers to be used.
// It returns a pointer to the Processor interface.
func NewProcessor(cfg config.CSV, hg web.HttpGateway, logger logs.Logger, workers int) Processor {
	return &processorImpl{
		csvConfig: cfg,
		gateway:   hg,
		logger:    logger,
		workers:   utils.Clamp(workers, 1, MAX_WORKERS),
	}
}

// Do performs the processing of a file specified by the given filePath.
// It creates a channel to receive the output from the mapCSV function and spawns multiple worker goroutines to process the output concurrently.
// Once all the workers have finished processing, it checks if there were any requests processed and logs a message if there were any errors.
// Finally, it resets the request, error, and lines counters and cancels the context.
func (this *processorImpl) Do(ctx context.Context, cancel func(), filePath string) {
	out := this.mapCSV(ctx, filePath)

	if out == nil {
		cancel()
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(this.workers)
	for i := 0; i < this.workers; i++ {
		go this.worker(ctx, wg, out)
	}

	go func() {
		wg.Wait()

		if reqCount.Load() > 0 {
			this.logger.Add(doneMessage(errCount.Load()))
		}
		reqCount.Store(0)
		errCount.Store(0)
		linesCount.Store(0)
		cancel()
	}()
}

func (this *processorImpl) worker(ctx context.Context, wg *sync.WaitGroup, out <-chan csvLineMap) {
	defer wg.Done()

Processing:
	for row := range out {
		select {
		case <-ctx.Done():
			this.logger.Add(cancelationMsg())
			break Processing
		default:
			response, err := this.gateway.Exec(row)
			reqCount.Add(1)
			if err != nil {
				errCount.Add(1)
				this.logger.Add(requestError(err.Error()))
			} else if response.StatusCode != http.StatusOK {
				errCount.Add(1)
				this.logger.Add(httpStatusError(row, response.StatusCode))
			}
			this.logger.WriteToFile(&RequestLine{URL: response.URL, Status: response.StatusCode, Body: response.Body, Error: err})
		}
	}
}

func (this *processorImpl) mapCSV(ctx context.Context, filePath string) <-chan csvLineMap {
	out := make(chan csvLineMap, this.workers)

	reader, file, err := newCSVReader(filePath, csvSep(this.csvConfig))
	if err != nil {
		this.logger.Add(csvError(err.Error()))
		return nil
	}

	headers, err := readCSVHeaders(reader)
	if err != nil {
		this.logger.Add(csvError(err.Error()))
		return nil
	}

	indexes := buildFilteredFieldIndex(headers, this.csvConfig.Fields)
	this.logger.Add(processingMessage(filepath.Base(filePath), this.workers))

	go func() {
		defer file.Close()
		defer close(out)
	Read:
		for {
			select {
			case <-ctx.Done():
				break Read
			default:
				record, err := reader.Read()
				if err == io.EOF {
					break Read
				}
				if err != nil {
					this.logger.Add(csvError(err.Error()))
					continue
				}
				linesCount.Add(1)
				out <- mapRow(headers, indexes, record)
			}
		}
	}()

	return out
}
