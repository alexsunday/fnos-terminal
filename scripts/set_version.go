package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	manifestPath := "package/manifest"
	uiPkgPath := "ui/package.json"

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", manifestPath, err)
		os.Exit(1)
	}
	content := string(data)

	versionRe := regexp.MustCompile(`(?m)^version=(.+)$`)
	matches := versionRe.FindStringSubmatch(content)
	currentVersion := "1.0.0"
	if len(matches) > 1 {
		currentVersion = strings.TrimSpace(matches[1])
	}

	targetVersion := currentVersion
	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) != "" {
		targetVersion = strings.TrimSpace(os.Args[1])
	}

	// Update package/manifest if changed
	if targetVersion != currentVersion {
		newContent := versionRe.ReplaceAllString(content, "version="+targetVersion)
		if err := os.WriteFile(manifestPath, []byte(newContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", manifestPath, err)
			os.Exit(1)
		}

		// Also update ui/package.json if it exists
		if pkgData, readErr := os.ReadFile(uiPkgPath); readErr == nil {
			var pkgMap map[string]interface{}
			if jsonErr := json.Unmarshal(pkgData, &pkgMap); jsonErr == nil {
				pkgMap["version"] = targetVersion
				if formatted, marshalErr := json.MarshalIndent(pkgMap, "", "  "); marshalErr == nil {
					_ = os.WriteFile(uiPkgPath, append(formatted, '\n'), 0644)
				}
			}
		}
	}

	// Output target version to stdout for shell capture
	fmt.Print(targetVersion)
}
