package paths

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

var ErrGlobalConfigNotFound = errors.New("global config not found")

type GlobalConfig struct {
	DefaultRoot string `toml:"default_root"`
}

func ExpandUserPath(path string) (string, error) {
	value := strings.TrimSpace(path)
	switch {
	case value == "":
		return "", nil
	case value == "~":
		return os.UserHomeDir()
	case strings.HasPrefix(value, "~/"), strings.HasPrefix(value, "~\\"):
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to resolve user home dir: %w", err)
		}
		return filepath.Join(home, value[2:]), nil
	default:
		return value, nil
	}
}

func GlobalConfigPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config dir: %w", err)
	}
	return filepath.Join(cfgDir, "gitshelf", "config.toml"), nil
}

func DefaultGlobalRoot() (string, error) {
	var base string
	if runtime.GOOS == "windows" {
		cfgDir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("failed to resolve user config dir: %w", err)
		}
		base = cfgDir
	} else {
		if xdgData := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdgData != "" {
			base = xdgData
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to resolve user home dir: %w", err)
			}
			base = filepath.Join(home, ".local", "share")
		}
	}

	root := filepath.Join(base, "gitshelf")
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("failed to make global root absolute: %w", err)
	}
	return abs, nil
}

func LoadGlobalConfig() (GlobalConfig, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return GlobalConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return GlobalConfig{}, ErrGlobalConfigNotFound
		}
		return GlobalConfig{}, fmt.Errorf("failed to read global config %s: %w", path, err)
	}

	var cfg GlobalConfig
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return GlobalConfig{}, fmt.Errorf("failed to parse global config %s: %w", path, err)
	}
	cfg.DefaultRoot = strings.TrimSpace(cfg.DefaultRoot)
	if cfg.DefaultRoot == "" {
		return GlobalConfig{}, fmt.Errorf("%s: default_root is required", path)
	}
	cfg.DefaultRoot, err = ExpandUserPath(cfg.DefaultRoot)
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("%s: failed to expand default_root: %w", path, err)
	}
	if !filepath.IsAbs(cfg.DefaultRoot) {
		abs, err := filepath.Abs(cfg.DefaultRoot)
		if err != nil {
			return GlobalConfig{}, fmt.Errorf("%s: failed to make default_root absolute: %w", path, err)
		}
		cfg.DefaultRoot = abs
	}
	return cfg, nil
}

func SaveGlobalConfig(cfg GlobalConfig) error {
	path, err := GlobalConfigPath()
	if err != nil {
		return err
	}
	cfg.DefaultRoot = strings.TrimSpace(cfg.DefaultRoot)
	if cfg.DefaultRoot == "" {
		return fmt.Errorf("default_root is required")
	}
	cfg.DefaultRoot, err = ExpandUserPath(cfg.DefaultRoot)
	if err != nil {
		return fmt.Errorf("failed to expand default_root: %w", err)
	}
	if !filepath.IsAbs(cfg.DefaultRoot) {
		abs, err := filepath.Abs(cfg.DefaultRoot)
		if err != nil {
			return fmt.Errorf("failed to make default_root absolute: %w", err)
		}
		cfg.DefaultRoot = abs
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("default_root = %q\n", cfg.DefaultRoot))

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create global config directory %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, "config.toml.tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp global config: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := tmp.Write(buf.Bytes()); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("failed to write temp global config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("failed to sync temp global config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("failed to close temp global config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("failed to rename global config: %w", err)
	}
	return nil
}
