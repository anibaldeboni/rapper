package processor

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/anibaldeboni/rapper/cli/messages"
	"github.com/anibaldeboni/rapper/internal/files"
	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/output"
	"github.com/anibaldeboni/rapper/internal/web"
)

type Processor interface {
	Do(ctx context.Context, cancel func(), filePath string)
}

type processorImpl struct {
	gateway      web.HttpGateway
	outputStream output.Stream
	logs         log.LogManager
	sep          rune
	fields       []string
}

func New(config files.AppConfig, outputFile string, lr log.LogManager) Processor {
	sep := strings.Trim(config.CSV.Separator, " ")

	if sep == "" {
		sep = ","
	}
	hg := web.NewHttpGateway(
		config.Token,
		config.Path.Method,
		config.Path.Template,
		config.Payload.Template,
	)

	return &processorImpl{
		gateway:      hg,
		outputStream: output.New(outputFile, lr),
		logs:         lr,
		sep:          []rune(sep)[0],
		fields:       config.CSV.Fields,
	}
}

func (p *processorImpl) Do(ctx context.Context, cancel func(), filePath string) {
	out := make(chan map[string]string)
	go p.mapCSV(ctx, cancel, filePath, out)
	go p.doRequests(ctx, cancel, out)
}

func (p *processorImpl) doRequests(ctx context.Context, cancel func(), out <-chan map[string]string) {
	var count, errCount int
	defer cancel()

Processing:
	for row := range out {
		select {
		case <-ctx.Done():
			p.logs.Add(messages.NewCancelationError(fmt.Sprintf("Processed %d lines", count)))
			break Processing
		default:
			response, err := p.gateway.Exec(row)
			count++
			if err != nil {
				errCount++
				p.logs.Add(messages.NewRequestError(err.Error()))
			} else if response.StatusCode != http.StatusOK {
				errCount++
				p.logs.Add(messages.NewHttpStatusError(row, response.StatusCode))
			}
			p.outputStream.Send(output.NewMessage(response.URL, response.StatusCode, err, response.Body))
		}
	}

	if count > 0 {
		p.logs.Add(messages.NewDoneMessage(errCount))
	}
}

func (p *processorImpl) mapCSV(ctx context.Context, cancel func(), filePath string, out chan<- map[string]string) {
	defer close(out)
	file, err := os.Open(filePath)
	if err != nil {
		p.logs.Add(messages.NewCsvError(err.Error()))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = p.sep
	headers, err := reader.Read()
	if err != nil {
		var e string
		if err == io.EOF {
			e = "No records found in the file\n"
		} else {
			e = err.Error()
		}
		p.logs.Add(messages.NewCsvError(e))
		cancel()
		return
	}

	headerIndexes := headerIndexes(headers, p.fields)
	p.logs.Add(messages.NewProcessingMessage(filepath.Base(filePath)))

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
				p.logs.Add(messages.NewCsvError(err.Error()))
				continue
			}

			row := make(map[string]string)
			for i, header := range headers {
				if _, ok := headerIndexes[header]; ok {
					row[header] = record[i]
				}
			}

			out <- row
		}
	}
}

func headerIndexes(headers []string, fields []string) map[string]int {
	var f []string
	indexes := make(map[string]int, len(fields))
	if len(fields) == 0 {
		f = headers
	} else {
		f = fields
	}
	for i, field := range f {
		indexes[field] = i
	}

	return indexes
}
