package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const DefaultVersion = "0.1.0"

func Get() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return DefaultVersion
	}

	version := info.Main.Version
	if version == "" || version == "(devel)" {
		version = DefaultVersion
	}

	var revision, modified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value
		}
	}

	if revision != "" {
		shortRev := revision
		if len(revision) > 7 {
			shortRev = revision[:7]
		}
		version = fmt.Sprintf("%s (%s)", version, shortRev)
		if modified == "true" {
			version += " (modified)"
		}
	}

	return version
}

func GetFull() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return DefaultVersion
	}

	var parts []string
	parts = append(parts, "Version: "+Get())

	for _, setting := range info.Settings {
		switch setting.Key {
		case "GOOS":
			parts = append(parts, "OS: "+setting.Value)
		case "GOARCH":
			parts = append(parts, "Arch: "+setting.Value)
		case "vcs.time":
			parts = append(parts, "Build time: "+setting.Value)
		}
	}

	return strings.Join(parts, "\n")
}
