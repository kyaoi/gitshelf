package shelf

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func LinkTasks(rootDir string, fromID, toID string, linkType LinkType) error {
	fromID = strings.TrimSpace(fromID)
	toID = strings.TrimSpace(toID)
	if fromID == "" || toID == "" {
		return fmt.Errorf("from/to は必須です")
	}
	if fromID == toID {
		return fmt.Errorf("from と to は同一にできません")
	}

	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return err
	}
	if err := cfg.ValidateLinkType(linkType); err != nil {
		return err
	}

	taskStore := NewTaskStore(rootDir)
	if _, err := taskStore.Get(fromID); err != nil {
		return fmt.Errorf("from タスクが存在しません: %s", fromID)
	}
	if _, err := taskStore.Get(toID); err != nil {
		return fmt.Errorf("to タスクが存在しません: %s", toID)
	}
	if linkType == LinkType("depends_on") {
		cycle, err := wouldCreateDependsOnCycle(rootDir, fromID, toID)
		if err != nil {
			return err
		}
		if cycle {
			return fmt.Errorf("depends_on cycle detected: %s -> %s", fromID, toID)
		}
	}

	edgeStore := NewEdgeStore(rootDir)
	return edgeStore.AddOutbound(fromID, Edge{
		To:   toID,
		Type: linkType,
	}, cfg.LinkTypes)
}

func UnlinkTasks(rootDir string, fromID, toID string, linkType LinkType) (bool, error) {
	fromID = strings.TrimSpace(fromID)
	toID = strings.TrimSpace(toID)
	if fromID == "" || toID == "" {
		return false, fmt.Errorf("from/to は必須です")
	}

	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return false, err
	}
	if err := cfg.ValidateLinkType(linkType); err != nil {
		return false, err
	}

	edgeStore := NewEdgeStore(rootDir)
	return edgeStore.RemoveOutbound(fromID, Edge{
		To:   toID,
		Type: linkType,
	})
}

func ListLinks(rootDir, taskID string) ([]Edge, []InboundEdge, error) {
	taskStore := NewTaskStore(rootDir)
	if _, err := taskStore.Get(taskID); err != nil {
		return nil, nil, fmt.Errorf("task が存在しません: %s", taskID)
	}

	edgeStore := NewEdgeStore(rootDir)
	outbound, err := edgeStore.ListOutbound(taskID)
	if err != nil {
		return nil, nil, err
	}
	inbound, err := edgeStore.FindInbound(taskID)
	if err != nil {
		return nil, nil, err
	}
	return outbound, inbound, nil
}

func ListTransitiveDependencies(rootDir, taskID string) ([]string, error) {
	taskStore := NewTaskStore(rootDir)
	if _, err := taskStore.Get(taskID); err != nil {
		return nil, fmt.Errorf("task が存在しません: %s", taskID)
	}
	adj, err := buildDependsOnAdjacency(rootDir)
	if err != nil {
		return nil, err
	}

	visited := map[string]struct{}{}
	stack := append([]string{}, adj[taskID]...)
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, ok := visited[cur]; ok {
			continue
		}
		visited[cur] = struct{}{}
		stack = append(stack, adj[cur]...)
	}

	result := make([]string, 0, len(visited))
	for id := range visited {
		result = append(result, id)
	}
	slices.Sort(result)
	return result, nil
}

func wouldCreateDependsOnCycle(rootDir, fromID, toID string) (bool, error) {
	adj, err := buildDependsOnAdjacency(rootDir)
	if err != nil {
		return false, err
	}
	seen := map[string]struct{}{}
	stack := append([]string{}, adj[toID]...)
	stack = append(stack, toID)
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cur == fromID {
			return true, nil
		}
		if _, ok := seen[cur]; ok {
			continue
		}
		seen[cur] = struct{}{}
		stack = append(stack, adj[cur]...)
	}
	return false, nil
}

func buildDependsOnAdjacency(rootDir string) (map[string][]string, error) {
	adj := map[string][]string{}
	entries, err := os.ReadDir(EdgesDir(rootDir))
	if err != nil {
		if os.IsNotExist(err) {
			return adj, nil
		}
		return nil, err
	}
	edgeStore := NewEdgeStore(rootDir)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		srcID := strings.TrimSuffix(entry.Name(), ".toml")
		outbound, err := edgeStore.ListOutbound(srcID)
		if err != nil {
			return nil, err
		}
		for _, edge := range outbound {
			if edge.Type != LinkType("depends_on") {
				continue
			}
			adj[srcID] = append(adj[srcID], edge.To)
		}
		slices.Sort(adj[srcID])
	}
	return adj, nil
}
