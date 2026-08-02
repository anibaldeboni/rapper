package processor

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
