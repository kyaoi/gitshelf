package cli

import (
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func loadOutputPreset(rootDir string, name string, command string) (shelf.OutputPreset, error) {
	if strings.TrimSpace(name) == "" {
		return shelf.OutputPreset{}, nil
	}
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return shelf.OutputPreset{}, err
	}
	preset, ok := cfg.OutputPresets[name]
	if !ok {
		return shelf.OutputPreset{}, fmt.Errorf("unknown output preset: %s", name)
	}
	if preset.Command != command {
		return shelf.OutputPreset{}, fmt.Errorf("output preset %s is for %s, not %s", name, preset.Command, command)
	}
	return preset, nil
}

func applyPresetString(current string, changed bool, preset string) string {
	if changed || strings.TrimSpace(preset) == "" {
		return current
	}
	return preset
}

func applyPresetInt(current int, changed bool, preset int) int {
	if changed || preset <= 0 {
		return current
	}
	return preset
}
