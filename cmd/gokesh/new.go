package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runNew(name, mdDir string) error {
	// sanitise: strip .md suffix if user included it
	name = strings.TrimSuffix(name, ".md")

	path := filepath.Join(mdDir, name+".md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	today := time.Now().Format("2006-01-02")
	content := fmt.Sprintf("---\ntitle: %q\ndate: %s\ndraft: true\n---\n\n", name, today)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	slog.Info("created", "file", path)
	fmt.Println(path)
	return nil
}
