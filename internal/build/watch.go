package build

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Watch polls source directories every second and rebuilds when any file changes.
// Config is reloaded from gokesh.toml on every rebuild so changes take effect
// without restarting the process. Blocks until the process is killed.
func Watch(outpath string, configPath string, templatesDir string) error {
	return WatchWithCallback(outpath, configPath, templatesDir, nil)
}

// WatchWithCallback is like Watch but calls onRebuild (if non-nil) after each
// successful build. Used by the serve command to trigger live reload.
func WatchWithCallback(outpath string, configPath string, templatesDir string, onRebuild func()) error {
	watched := []string{"./markdown/", "./templates/", "./styles/", "./static/", "./data/", "./gokesh.toml"}

	slog.Info("watching for changes — press Ctrl+C to stop")

	// initial build
	lastBuild := time.Now()
	if err := fullBuild(outpath, configPath, templatesDir); err != nil {
		slog.Error("build failed", "error", err)
	} else if onRebuild != nil {
		onRebuild()
	}

	for {
		time.Sleep(time.Second)
		if latestMtime(watched...).After(lastBuild) {
			lastBuild = time.Now()
			slog.Info("change detected, rebuilding")
			if err := fullBuild(outpath, configPath, templatesDir); err != nil {
				slog.Error("build failed", "error", err)
			} else if onRebuild != nil {
				onRebuild()
			}
		}
	}
}

func fullBuild(outpath string, configPath string, templatesDir string) error {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if err := CopyStyles("./styles/", outpath); err != nil {
		return err
	}
	if err := CopyStatic("./static/", outpath); err != nil {
		return err
	}
	if err := BuildAllIncremental("./markdown/", outpath, cfg, templatesDir); err != nil {
		return err
	}
	pages := CollectPages("./markdown/", cfg.BaseURL)
	if err := GenerateRSS(pages, outpath, cfg.SiteTitle, cfg.BaseURL); err != nil {
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
