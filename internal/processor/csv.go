package processor

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/utils"
)

func readCSVHeaders(reader *csv.Reader) ([]string, error) {
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			err = errors.New("No records found in the file\n")
		}
		return nil, fmt.Errorf("Error reading headers: %w", err)
	}

	return headers, nil
}

func newCSVReader(filePath string, sep rune) (*csv.Reader, *os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error opening file: %w", err)
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

func buildFilteredFieldIndex(csvHeaders []string, fieldsToFilter []string) map[string]int {
	indexes := make(map[string]int, len(csvHeaders))
	if len(fieldsToFilter) == 0 {
		fieldsToFilter = csvHeaders
	}
	for i, field := range fieldsToFilter {
		indexes[field] = i
	}

	return indexes
}

func csvSep(cfg config.CSVConfig) rune {
	sep := strings.Trim(cfg.Separator, " ")
	if utils.IsZero(sep) {
		sep = ","
	}
	return rune(sep[0])
}
