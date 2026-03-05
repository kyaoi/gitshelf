package shelf

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

type Edge struct {
	To   string   `toml:"to"`
	Type LinkType `toml:"type"`
}

type InboundEdge struct {
	From string
	To   string
	Type LinkType
}

type edgeFile struct {
	Edges []Edge `toml:"edge"`
}

func ParseEdgesTOML(data []byte) ([]Edge, error) {
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}

	var f edgeFile
	if _, err := toml.Decode(string(data), &f); err != nil {
		return nil, fmt.Errorf("failed to parse edges TOML: %w", err)
	}
	edges := normalizeEdges(f.Edges)
	return edges, nil
}

func FormatEdgesTOML(edges []Edge) []byte {
	normalized := normalizeEdges(edges)

	var buf bytes.Buffer
	for i, edge := range normalized {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("[[edge]]\n")
		buf.WriteString(fmt.Sprintf("to = %q\n", edge.To))
		buf.WriteString(fmt.Sprintf("type = %q\n", edge.Type))
	}
	if buf.Len() > 0 {
		buf.WriteString("\n")
	}
	return buf.Bytes()
}

func normalizeEdges(edges []Edge) []Edge {
	filtered := make([]Edge, 0, len(edges))
	seen := make(map[string]struct{}, len(edges))
	for _, edge := range edges {
		edge.To = strings.TrimSpace(edge.To)
		edge.Type = LinkType(strings.TrimSpace(string(edge.Type)))
		if edge.To == "" || edge.Type == "" {
			continue
		}
		key := string(edge.Type) + "\x00" + edge.To
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		filtered = append(filtered, edge)
	}

	slices.SortFunc(filtered, func(a, b Edge) int {
		if a.To < b.To {
			return -1
		}
		if a.To > b.To {
			return 1
		}
		if a.Type < b.Type {
			return -1
		}
		if a.Type > b.Type {
			return 1
		}
		return 0
	})
	return filtered
}

func ValidateLinkType(linkType LinkType, allowed []LinkType) error {
	if strings.TrimSpace(string(linkType)) == "" {
		return fmt.Errorf("link type is required")
	}
	for _, kind := range allowed {
		if kind == linkType {
			return nil
		}
	}
	return fmt.Errorf("unknown link type: %s", linkType)
}
