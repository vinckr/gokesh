package build

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Watch polls source directories every second and rebuilds when any file changes.
// Blocks until the process is killed.
func Watch(outpath string, cfg Config, templatesDir string) error {
	watched := []string{"./markdown/", "./templates/", "./styles/", "./gokesh.toml"}

	slog.Info("watching for changes — press Ctrl+C to stop")

	// initial build
	lastBuild := time.Now()
	if err := fullBuild(outpath, cfg, templatesDir); err != nil {
		slog.Error("build failed", "error", err)
	}

	for {
		time.Sleep(time.Second)
		if latestMtime(watched...).After(lastBuild) {
			lastBuild = time.Now()
			slog.Info("change detected, rebuilding")
			if err := fullBuild(outpath, cfg, templatesDir); err != nil {
				slog.Error("build failed", "error", err)
			}
		}
	}
}

func fullBuild(outpath string, cfg Config, templatesDir string) error {
	if err := CopyStyles("./styles/", outpath); err != nil {
		return err
	}
	if err := BuildAll("./markdown/", outpath, cfg, templatesDir); err != nil {
		return err
	}
	return GenerateSitemap(outpath, cfg.BaseURL)
}

func latestMtime(paths ...string) time.Time {
	var latest time.Time
	for _, p := range paths {
		_ = filepath.Walk(p, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if t := info.ModTime(); t.After(latest) {
				latest = t
			}
			return nil
		})
	}
	return latest
}
