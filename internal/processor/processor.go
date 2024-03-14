package processor

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/anibaldeboni/rapper/internal"
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/execlog"
	"github.com/anibaldeboni/rapper/internal/filelogger"
	"github.com/anibaldeboni/rapper/internal/web"
)

var (
	_           Processor = (*processorImpl)(nil)
	reqCount    atomic.Uint64
	errCount    atomic.Uint64
	linesCount  atomic.Uint64
	MAX_WORKERS = 5
)

//go:generate mockgen -destination mock/processor_mock.go github.com/anibaldeboni/rapper/internal/processor Processor
type Processor interface {
	Do(ctx context.Context, cancel func(), filePath string)
}

type processorImpl struct {
	csvConfig  config.CSV
	gateway    web.HttpGateway
	fileLogger filelogger.FileLogger
	logManager execlog.Manager
	workers    int
}

func New(cfg config.CSV, hg web.HttpGateway, fileLogger filelogger.FileLogger, logManager execlog.Manager, workers int) Processor {
	return &processorImpl{
		csvConfig:  cfg,
		gateway:    hg,
		fileLogger: fileLogger,
		logManager: logManager,
		workers:    internal.Clamp(workers, 1, MAX_WORKERS),
	}
}

func (this *processorImpl) Do(ctx context.Context, cancel func(), filePath string) {
	out := make(chan map[string]string)
	go this.mapCSV(ctx, cancel, filePath, out)

	wg := &sync.WaitGroup{}

	for i := 0; i < this.workers; i++ {
		wg.Add(1)
		go this.worker(ctx, wg, out)
	}

	go func() {
		wg.Wait()

		if reqCount.Load() > 0 {
			this.logManager.Add(doneMessage(errCount.Load()))
		}
		reqCount.Store(0)
		errCount.Store(0)
		linesCount.Store(0)
		cancel()
	}()
}

func (this *processorImpl) worker(ctx context.Context, wg *sync.WaitGroup, out <-chan map[string]string) {
	defer wg.Done()

Processing:
	for row := range out {
		select {
		case <-ctx.Done():
			this.logManager.Add(cancelationMsg())
			break Processing
		default:
			response, err := this.gateway.Exec(row)
			reqCount.Add(1)
			if err != nil {
				errCount.Add(1)
				this.logManager.Add(requestError(err.Error()))
			} else if response.StatusCode != http.StatusOK {
				errCount.Add(1)
				this.logManager.Add(httpStatusError(row, response.StatusCode))
			}
			this.fileLogger.Write(filelogger.NewLine(response.URL, response.StatusCode, err, response.Body))
		}
	}
}

func (this *processorImpl) mapCSV(ctx context.Context, cancel func(), filePath string, out chan<- map[string]string) {
	defer close(out)

	reader, file, err := buildCSVReader(filePath, csvSep(this.csvConfig))
	if err != nil {
		this.logManager.Add(csvError(err.Error()))
		cancel()
		return
	}
	defer file.Close()

	headers, err := getCSVHeaders(reader)
	if err != nil {
		this.logManager.Add(csvError(err.Error()))
		cancel()
		return
	}

	indexes := headerIndexes(headers, this.csvConfig.Fields)
	this.logManager.Add(processingMessage(filepath.Base(filePath), this.workers))

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
				this.logManager.Add(csvError(err.Error()))
				continue
			}
			linesCount.Add(1)
			out <- mapRow(headers, indexes, record)
		}
	}
}
