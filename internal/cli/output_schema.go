package cli

import (
	"fmt"
	"strings"
)

type outputSchema string

const (
	outputSchemaV1 outputSchema = "v1"
	outputSchemaV2 outputSchema = "v2"
)

func parseOutputSchema(value string) (outputSchema, error) {
	switch outputSchema(strings.TrimSpace(value)) {
	case "", outputSchemaV1:
		return outputSchemaV1, nil
	case outputSchemaV2:
		return outputSchemaV2, nil
	default:
		return "", fmt.Errorf("unknown --schema: %s (allowed: v1|v2)", value)
	}
}
