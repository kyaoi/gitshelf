package versioninfo

import (
	"runtime/debug"
	"strings"
)

const develVersion = "(devel)"

func Resolve(explicit string) string {
	return resolve(explicit, readBuildInfo())
}

func resolve(explicit string, info *debug.BuildInfo) string {
	if value := strings.TrimSpace(explicit); value != "" {
		return value
	}
	if info != nil {
		if value := strings.TrimSpace(info.Main.Version); value != "" && value != develVersion {
			return value
		}
		if revision := shortRevision(buildSetting(info, "vcs.revision")); revision != "" {
			version := "dev+" + revision
			if buildSetting(info, "vcs.modified") == "true" {
				version += "-dirty"
			}
			return version
		}
	}
	return "dev"
}

func readBuildInfo() *debug.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return info
}

func buildSetting(info *debug.BuildInfo, key string) string {
	if info == nil {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}

func shortRevision(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 7 {
		return value[:7]
	}
	return value
}
