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
	blocking := cfg.BlockingLinkType()
	if linkType == blocking {
		ancestor, err := wouldLinkBlockingToAncestor(rootDir, fromID, toID)
		if err != nil {
			return err
		}
		if ancestor {
			return fmt.Errorf("%s cannot target an ancestor task: %s -> %s", blocking, fromID, toID)
		}
		cycle, err := wouldCreateDependsOnCycle(rootDir, fromID, toID, blocking)
		if err != nil {
			return err
		}
		if cycle {
			return fmt.Errorf("%s cycle detected: %s -> %s", blocking, fromID, toID)
		}
	}

	edgeStore := NewEdgeStore(rootDir)
	return edgeStore.AddOutbound(fromID, Edge{
		To:   toID,
		Type: linkType,
	}, cfg.LinkTypes.Names)
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
	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return nil, err
	}
	adj, err := buildDependsOnAdjacency(rootDir, cfg.BlockingLinkType())
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

func wouldCreateDependsOnCycle(rootDir, fromID, toID string, blocking LinkType) (bool, error) {
	adj, err := buildDependsOnAdjacency(rootDir, blocking)
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

func wouldLinkBlockingToAncestor(rootDir, fromID, toID string) (bool, error) {
	taskStore := NewTaskStore(rootDir)
	tasks, err := taskStore.List()
	if err != nil {
		return false, err
	}
	parentByID := make(map[string]string, len(tasks))
	for _, task := range tasks {
		parentByID[task.ID] = strings.TrimSpace(task.Parent)
	}
	seen := map[string]struct{}{fromID: {}}
	current := parentByID[fromID]
	for current != "" {
		if current == toID {
			return true, nil
		}
		if _, ok := seen[current]; ok {
			break
		}
		seen[current] = struct{}{}
		current = parentByID[current]
	}
	return false, nil
}

func buildDependsOnAdjacency(rootDir string, blocking LinkType) (map[string][]string, error) {
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
			if edge.Type != blocking {
				continue
			}
			adj[srcID] = append(adj[srcID], edge.To)
		}
		slices.Sort(adj[srcID])
	}
	return adj, nil
}
