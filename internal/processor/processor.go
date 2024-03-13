package processor

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/anibaldeboni/rapper/cli/messages"
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/output"
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
	csvConfig    config.CSV
	gateway      web.HttpGateway
	outputStream output.Stream
	logs         log.LogManager
}

func New(cfg config.CSV, hg web.HttpGateway, outputFile string, lr log.LogManager) Processor {
	return &processorImpl{
		csvConfig:    cfg,
		gateway:      hg,
		outputStream: output.New(outputFile, lr),
		logs:         lr,
	}
}

func (this *processorImpl) Do(ctx context.Context, cancel func(), filePath string) {
	out := make(chan map[string]string)
	go this.mapCSV(ctx, cancel, filePath, out)

	wg := &sync.WaitGroup{}

	// for i := 0; i < MAX_WORKERS; i++ {
	wg.Add(1)
	go this.reqWorker(ctx, wg, out)
	// }
	go func() {
		wg.Wait()

		if reqCount.Load() > 0 {
			this.logs.Add(messages.NewDoneMessage(errCount.Load(), linesCount.Load()))
		}
		reqCount.Store(0)
		errCount.Store(0)
		linesCount.Store(0)
		cancel()
	}()
}

func (this *processorImpl) reqWorker(ctx context.Context, wg *sync.WaitGroup, out <-chan map[string]string) {
	defer wg.Done()

Processing:
	for row := range out {
		select {
		case <-ctx.Done():
			this.logs.Add(messages.NewCancelationError(fmt.Sprintf("Read %d lines and executed %d requests", linesCount.Load(), reqCount.Load())))
			break Processing
		default:
			response, err := this.gateway.Exec(row)
			reqCount.Add(1)
			if err != nil {
				errCount.Add(1)
				this.logs.Add(messages.NewRequestError(err.Error()))
			} else if response.StatusCode != http.StatusOK {
				errCount.Add(1)
				this.logs.Add(messages.NewHttpStatusError(row, response.StatusCode))
			}
			this.outputStream.Send(output.NewMessage(response.URL, response.StatusCode, err, response.Body))
		}
	}
}

func csvSep(cfg config.CSV) rune {
	sep := strings.Trim(cfg.Separator, " ")
	if sep == "" {
		sep = ","
	}
	return rune(sep[0])
}

func (this *processorImpl) mapCSV(ctx context.Context, cancel func(), filePath string, out chan<- map[string]string) {
	defer close(out)

	reader, file, err := buildCSVReader(filePath, csvSep(this.csvConfig))
	defer file.Close()
	if err != nil {
		this.logs.Add(messages.NewCsvError(err.Error()))
		cancel()
		return
	}

	headers, err := getCSVHeaders(reader)
	if err != nil {
		this.logs.Add(messages.NewCsvError(err.Error()))
		cancel()
		return
	}

	indexes := headerIndexes(headers, this.csvConfig.Fields)
	this.logs.Add(messages.NewProcessingMessage(filepath.Base(filePath)))

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
				this.logs.Add(messages.NewCsvError(err.Error()))
				continue
			}
			linesCount.Add(1)
			out <- mapRow(headers, indexes, record)
		}
	}
}

func getCSVHeaders(reader *csv.Reader) ([]string, error) {
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			err = errors.New("No records found in the file\n")
		}
		return nil, err
	}

	return headers, nil
}

func buildCSVReader(filePath string, sep rune) (*csv.Reader, *os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	reader := csv.NewReader(file)
	reader.Comma = sep

	return reader, file, nil
}

func mapRow(headers []string, indexes map[string]int, record []string) map[string]string {
	row := make(map[string]string)
	for i, header := range headers {
		if _, ok := indexes[header]; ok {
			row[header] = record[i]
		}
	}
	return row
}

func headerIndexes(headers []string, fields []string) map[string]int {
	indexes := make(map[string]int, len(headers))
	if len(fields) == 0 {
		fields = headers
	}
	for i, field := range fields {
		indexes[field] = i
	}

	return indexes
}
