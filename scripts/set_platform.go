package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run scripts/set_platform.go <x86|arm|all>\n")
		os.Exit(1)
	}
	platform := strings.TrimSpace(os.Args[1])
	manifestPath := "package/manifest"

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", manifestPath, err)
		os.Exit(1)
	}

	re := regexp.MustCompile(`(?m)^platform=.+$`)
	if !re.Match(data) {
		fmt.Fprintf(os.Stderr, "platform field not found in %s\n", manifestPath)
		os.Exit(1)
	}

	newData := re.ReplaceAll(data, []byte("platform="+platform))
	if err := os.WriteFile(manifestPath, newData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", manifestPath, err)
		os.Exit(1)
	}
	fmt.Printf("Manifest platform updated to: %s\n", platform)
}
