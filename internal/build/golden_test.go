package build

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update golden files")

// fixedNow is a stable timestamp so golden files never need updating due to the year changing.
var fixedNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

var testConfig = Config{
	Author:      "testauthor",
	SiteTitle:   "Test Site",
	BaseURL:     "https://example.com",
	Description: "A test site",
}

const testTemplatesDir = "../../templates/"

func TestGoldenBuild(t *testing.T) {
	markdownRoot := "../../markdown"

	err := filepath.Walk(markdownRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}

		dir := filepath.Dir(path) + string(filepath.Separator)
		fileName := filepath.Base(path)
		rel, _ := filepath.Rel(markdownRoot, path)
		base := strings.TrimSuffix(filepath.Base(rel), ".md")
		var goldenPath, builtPath string
		if base == "index" {
			goldenPath = filepath.Join("testdata", "golden", filepath.Dir(rel), "index.html")
			builtPath = "index.html"
		} else {
			goldenPath = filepath.Join("testdata", "golden", strings.TrimSuffix(rel, ".md"), "index.html")
			builtPath = filepath.Join(base, "index.html")
		}

		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			outDir := t.TempDir()
			if err := BuildPageAt(fileName, dir, outDir+string(filepath.Separator), fixedNow, testConfig, testTemplatesDir); err != nil {
				t.Fatalf("BuildPageAt: %v", err)
			}

			actual, err := os.ReadFile(filepath.Join(outDir, builtPath))
			if err != nil {
				t.Fatalf("built output missing for %s: %v", path, err)
			}

			if *update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("creating golden dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
					t.Fatalf("writing golden file: %v", err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file missing for %s — run: go test ./internal/build/ -update\nerr: %v", path, err)
			}

			if string(actual) != string(golden) {
				t.Errorf("output mismatch for %s\n\ngot:\n%s\n\nwant:\n%s", path, actual, golden)
			}
		})

		return nil
	})

	if err != nil {
		t.Fatalf("walking markdown dir: %v", err)
	}
}
