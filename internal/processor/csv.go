package processor

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/anibaldeboni/rapper/internal/config"
)

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

func csvSep(cfg config.CSV) rune {
	sep := strings.Trim(cfg.Separator, " ")
	if sep == "" {
		sep = ","
	}
	return rune(sep[0])
}
