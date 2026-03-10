package cli

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
)

type outputFieldRecord interface {
	TSVFields() map[string]string
}

func renderCSV[T outputFieldRecord](records []T, fields []string, includeHeader bool) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if includeHeader {
		if err := writer.Write(fields); err != nil {
			return "", err
		}
	}
	for _, record := range records {
		rowMap := record.TSVFields()
		row := make([]string, 0, len(fields))
		for _, field := range fields {
			row = append(row, rowMap[field])
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderJSONL[T any](records []T) (string, error) {
	var buf bytes.Buffer
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			return "", err
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	return buf.String(), nil
}
