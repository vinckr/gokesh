package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func runClean(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("output directory does not exist, nothing to clean", "dir", outDir)
			return nil
		}
		return fmt.Errorf("reading %s: %w", outDir, err)
	}

	// count files for confirmation prompt
	count := 0
	_ = filepath.Walk(outDir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	_ = entries

	fmt.Printf("  Delete %s (%d files)? [y/N]: ", outDir, count)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
		fmt.Println("  aborted")
		return nil
	}

	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("removing %s: %w", outDir, err)
	}
	slog.Info("deleted", "dir", outDir)
	return nil
}
