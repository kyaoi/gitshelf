package shelf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type EdgeStore struct {
	rootDir string
}

func NewEdgeStore(rootDir string) *EdgeStore {
	return &EdgeStore{rootDir: rootDir}
}

func (s *EdgeStore) ListOutbound(srcID string) ([]Edge, error) {
	path := s.edgePath(srcID)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read edge file %s: %w", path, err)
	}

	edges, err := ParseEdgesTOML(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return edges, nil
}

func (s *EdgeStore) SetOutbound(srcID string, edges []Edge) error {
	path := s.edgePath(srcID)
	data := FormatEdgesTOML(edges)
	return atomicWriteFile(path, data, 0o644)
}

func (s *EdgeStore) AddOutbound(srcID string, edge Edge, allowedLinkTypes []LinkType) error {
	if err := ValidateLinkType(edge.Type, allowedLinkTypes); err != nil {
		return err
	}
	if strings.TrimSpace(edge.To) == "" {
		return fmt.Errorf("edge destination is required")
	}

	edges, err := s.ListOutbound(srcID)
	if err != nil {
		return err
	}
	edges = append(edges, edge)
	return s.SetOutbound(srcID, edges)
}

func (s *EdgeStore) RemoveOutbound(srcID string, edge Edge) (bool, error) {
	edges, err := s.ListOutbound(srcID)
	if err != nil {
		return false, err
	}

	before := len(edges)
	filtered := make([]Edge, 0, before)
	for _, item := range edges {
		if item.To == edge.To && item.Type == edge.Type {
			continue
		}
		filtered = append(filtered, item)
	}

	if len(filtered) == before {
		return false, nil
	}
	if err := s.SetOutbound(srcID, filtered); err != nil {
		return false, err
	}
	return true, nil
}

func (s *EdgeStore) FindInbound(taskID string) ([]InboundEdge, error) {
	entries, err := os.ReadDir(EdgesDir(s.rootDir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read edges directory: %w", err)
	}

	inbound := make([]InboundEdge, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		srcID := strings.TrimSuffix(entry.Name(), ".toml")
		outbound, err := s.ListOutbound(srcID)
		if err != nil {
			return nil, err
		}
		for _, edge := range outbound {
			if edge.To == taskID {
				inbound = append(inbound, InboundEdge{
					From: srcID,
					To:   edge.To,
					Type: edge.Type,
				})
			}
		}
	}
	slices.SortFunc(inbound, func(a, b InboundEdge) int {
		if a.From < b.From {
			return -1
		}
		if a.From > b.From {
			return 1
		}
		if a.Type < b.Type {
			return -1
		}
		if a.Type > b.Type {
			return 1
		}
		if a.To < b.To {
			return -1
		}
		if a.To > b.To {
			return 1
		}
		return 0
	})
	return inbound, nil
}

func (s *EdgeStore) edgePath(srcID string) string {
	return filepath.Join(EdgesDir(s.rootDir), srcID+".toml")
}
