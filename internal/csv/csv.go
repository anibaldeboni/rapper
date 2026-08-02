package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/anibaldeboni/rapper/internal/config"
)

type Preview struct {
	Headers []string
	Rows    [][]string
}

func Separator(cfg config.CSVConfig) rune {
	separator := strings.Trim(cfg.Separator, " ")
	if separator == "" {
		return ','
	}
	return rune(separator[0])
}

func Open(filePath string, separator rune) (*csv.Reader, *os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening file: %w", err)
	}

	reader := csv.NewReader(file)
	reader.Comma = separator

	return reader, file, nil
}

func Headers(reader *csv.Reader) ([]string, error) {
	headers, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = errors.New("no records found in the file")
		}
		return nil, fmt.Errorf("error reading headers: %w", err)
	}

	return headers, nil
}

func ReadPreview(filePath string, separator rune, maxRows int) (Preview, error) {
	reader, file, err := Open(filePath, separator)
	if err != nil {
		return Preview{}, err
	}
	defer file.Close()

	headers, err := Headers(reader)
	if err != nil {
		return Preview{}, err
	}

	preview := Preview{
		Headers: headers,
		Rows:    make([][]string, 0, maxRows),
	}
	for range maxRows {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return Preview{}, fmt.Errorf("error reading rows: %w", err)
		}
		preview.Rows = append(preview.Rows, record)
	}

	return preview, nil
}
