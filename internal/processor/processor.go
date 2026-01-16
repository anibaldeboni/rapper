package processor

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/anibaldeboni/rapper/internal/web"
)

var (
	_          Processor = (*processorImpl)(nil)
	reqCount   atomic.Uint64
	errCount   atomic.Uint64
	linesCount atomic.Uint64
	MaxWorkers = runtime.NumCPU()
)

// Metrics holds real-time processing metrics
type Metrics struct {
	TotalRequests   uint64
	SuccessRequests uint64
	ErrorRequests   uint64
	LinesProcessed  uint64
	ActiveWorkers   int
	RequestsPerSec  float64
	StartTime       time.Time
	IsProcessing    bool
}

//go:generate mockgen -destination mock/processor_mock.go github.com/anibaldeboni/rapper/internal/processor Processor
type Processor interface {
	Do(ctx context.Context, filePath string) (context.Context, context.CancelFunc)

	// GetMetrics returns current processing metrics
	GetMetrics() Metrics

	// SetWorkers dynamically adjusts the number of workers (only affects next processing)
	SetWorkers(n int)

	// GetWorkerCount returns the current worker count
	GetWorkerCount() int
}

type csvLineMap map[string]string

type processorImpl struct {
	gateway      web.HttpGateway
	logger       logs.Logger
	csvConfig    config.CSVConfig
	workers      int
	mu           sync.RWMutex
	startTime    time.Time
	isProcessing bool
}

// NewProcessor creates a new instance of the Processor interface.
// It takes in the following parameters:
// - cfg: The CSV configuration.
// - hg: The HTTP gateway.
// - logger: The logger.
// - workers: The number of workers to be used.
// It returns a pointer to the Processor interface.
func NewProcessor(cfg config.CSVConfig, hg web.HttpGateway, logger logs.Logger, workers int) Processor {
	return &processorImpl{
		csvConfig: cfg,
		gateway:   hg,
		logger:    logger,
		workers:   utils.Clamp(workers, 1, MaxWorkers),
	}
}

// Do performs the processing of a file specified by the given filePath.
// It creates a channel to receive the output from the mapCSV function and spawns multiple worker goroutines to process the output concurrently.
// Once all the workers have finished processing, it checks if there were any requests processed and logs a message if there were any errors.
// Finally, it resets the request, error, and lines counters and cancels the context.
func (p *processorImpl) Do(ctx context.Context, filePath string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	rows := p.mapCSV(ctx, filePath)

	if rows == nil {
		cancel()
		return nil, nil
	}

	// Mark processing as started
	p.mu.Lock()
	p.startTime = time.Now()
	p.isProcessing = true
	p.mu.Unlock()

	wg := &sync.WaitGroup{}
	wg.Add(p.workers)
	for i := 0; i < p.workers; i++ {
		go p.worker(ctx, wg, rows)
	}

	go func() {
		wg.Wait()

		// Mark processing as finished
		p.mu.Lock()
		p.isProcessing = false
		p.mu.Unlock()

		if reqCount.Load() > 0 {
			p.logger.Add(doneMessage(errCount.Load()))
		}
		reqCount.Store(0)
		errCount.Store(0)
		linesCount.Store(0)
		cancel()
	}()

	return ctx, cancel
}

func (p *processorImpl) worker(ctx context.Context, wg *sync.WaitGroup, rows <-chan csvLineMap) {
	defer wg.Done()

requests:
	for row := range rows {
		select {
		case <-ctx.Done():
			p.logger.Add(cancelationMsg())
			break requests
		default:
			res, err := p.gateway.Exec(ctx, row)
			reqCount.Add(1)
			if err != nil {
				errCount.Add(1)
				p.logger.Add(requestError(err.Error()))
			} else if res.StatusCode != http.StatusOK {
				errCount.Add(1)
				p.logger.Add(httpStatusError(row, res.StatusCode))
			}
			p.logger.WriteToFile(&RequestLine{URL: res.URL, Status: res.StatusCode, Body: res.Body, Error: err})
		}
	}
}

func (p *processorImpl) mapCSV(ctx context.Context, filePath string) <-chan csvLineMap {
	rows := make(chan csvLineMap, p.workers)

	reader, file, err := newCSVReader(filePath, csvSep(p.csvConfig))
	if err != nil {
		p.logger.Add(csvError(err.Error()))
		return nil
	}

	headers, err := readCSVHeaders(reader)
	if err != nil {
		p.logger.Add(csvError(err.Error()))
		return nil
	}

	indexes := buildFilteredFieldIndex(headers, p.csvConfig.Fields)
	p.logger.Add(processingMessage(filepath.Base(filePath), p.workers))

	go func() {
		defer file.Close()
		defer close(rows)

	read:
		for {
			select {
			case <-ctx.Done():
				break read
			default:
				record, err := reader.Read()
				if err == io.EOF {
					break read
				}
				if err != nil {
					p.logger.Add(csvError(err.Error()))
					continue
				}
				linesCount.Add(1)
				rows <- mapRow(headers, indexes, record)
			}
		}
	}()

	return rows
}

// GetMetrics returns current processing metrics
func (p *processorImpl) GetMetrics() Metrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalReq := reqCount.Load()
	errReq := errCount.Load()
	successReq := totalReq - errReq

	var reqPerSec float64
	if p.isProcessing && !p.startTime.IsZero() {
		elapsed := time.Since(p.startTime).Seconds()
		if elapsed > 0 {
			reqPerSec = float64(totalReq) / elapsed
		}
	}

	return Metrics{
		TotalRequests:   totalReq,
		SuccessRequests: successReq,
		ErrorRequests:   errReq,
		LinesProcessed:  linesCount.Load(),
		ActiveWorkers:   p.workers,
		RequestsPerSec:  reqPerSec,
		StartTime:       p.startTime,
		IsProcessing:    p.isProcessing,
	}
}

// SetWorkers dynamically adjusts the number of workers
// Note: This only affects the next processing run, not current processing
func (p *processorImpl) SetWorkers(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.workers = utils.Clamp(n, 1, MaxWorkers)
}

// GetWorkerCount returns the current configured worker count
func (p *processorImpl) GetWorkerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.workers
}
