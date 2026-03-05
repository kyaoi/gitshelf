package shelf

import (
	"fmt"
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
