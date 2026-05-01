package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinckr/gokesh/internal/build"
	gokesh "github.com/vinckr/gokesh"
)

func runInit() error {
	if err := setupConfig(); err != nil {
		return err
	}
	if err := copyEmbedDir("templates", "templates-examples", true); err != nil {
		return err
	}
	if err := setupCSS(); err != nil {
		return err
	}
	if err := copyEmbedDir("styles", "styles-examples", true); err != nil {
		return err
	}
	readme, err := gokesh.FS.ReadFile("README.md")
	if err != nil {
		return err
	}
	if err := os.WriteFile("README.md", readme, 0644); err != nil {
		return err
	}
	slog.Info("wrote README.md")

	cfg, err := build.LoadConfig("./gokesh.toml")
	if err != nil {
		return err
	}
	if err := os.MkdirAll("./public", 0755); err != nil {
		return err
	}
	if err := writeHeaders("./public/_headers"); err != nil {
		return err
	}
	if err := writeWebManifest("./public/site.webmanifest", cfg.SiteTitle); err != nil {
		return err
	}
	return nil
}

func setupCSS() error {
	useTailwind := promptYN("Use Tailwind CSS?")
	if !useTailwind {
		return nil
	}

	const headerPath = "./templates/header.tmpl"
	data, err := os.ReadFile(headerPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", headerPath, err)
	}

	const cdnTag = `    <script src="https://unpkg.com/@tailwindcss/browser@4"></script>`
	if strings.Contains(string(data), "tailwindcss/browser") {
		slog.Info("Tailwind CDN already present in header.tmpl, skipping")
		return nil
	}

	updated := strings.Replace(string(data), "</head>", cdnTag+"\n</head>", 1)
	if err := os.WriteFile(headerPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", headerPath, err)
	}
	slog.Info("added Tailwind CDN script to templates/header.tmpl")
	return nil
}


func promptYN(label string) bool {
	fmt.Printf("  %s [y/N]: ", label)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	val := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return val == "y" || val == "yes"
}

func setupConfig() error {
	if _, err := os.Stat("gokesh.toml"); err == nil {
		slog.Info("gokesh.toml already exists, skipping config setup")
		return nil
	}
	fmt.Println("Setting up gokesh.toml — press Enter to accept defaults.")
	author := prompt("Author", "")
	siteTitle := prompt("Site title", "")
	baseURL := prompt("Base URL (e.g. https://example.com)", "")
	description := prompt("Description", "")

	content := fmt.Sprintf(
		"author      = %q\nsite_title  = %q\nbase_url    = %q\ndescription = %q\n",
		author, siteTitle, baseURL, description,
	)
	if err := os.WriteFile("gokesh.toml", []byte(content), 0644); err != nil {
		return fmt.Errorf("writing gokesh.toml: %w", err)
	}
	slog.Info("wrote gokesh.toml")
	return nil
}

func prompt(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	val := strings.TrimSpace(scanner.Text())
	if val == "" {
		return defaultVal
	}
	return val
}

func writeHeaders(path string) error {
	const headers = `/*
  X-Frame-Options: SAMEORIGIN
  X-Content-Type-Options: nosniff
  X-XSS-Protection: 1; mode=block
  Referrer-Policy: strict-origin-when-cross-origin
  Content-Security-Policy: default-src 'self'; img-src 'self' data:; font-src 'self'; style-src 'self' 'unsafe-inline'
  Permissions-Policy: accelerometer=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), midi=(), payment=(), usb=()
`
	if err := os.WriteFile(path, []byte(headers), 0644); err != nil {
		return fmt.Errorf("writing _headers: %w", err)
	}
	slog.Info("wrote", "file", path)
	return nil
}

func writeWebManifest(path, siteTitle string) error {
	content := fmt.Sprintf(`{
  "name": %q,
  "short_name": %q,
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#000000",
  "icons": [
    { "src": "/img/icon-192.png", "type": "image/png", "sizes": "192x192" },
    { "src": "/img/icon-512.png", "type": "image/png", "sizes": "512x512" }
  ]
}
`, siteTitle, siteTitle)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing site.webmanifest: %w", err)
	}
	slog.Info("wrote", "file", path)
	return nil
}

// copyEmbedDir copies src from the embedded FS into dst on disk.
// If initLive is true and the live dir does not exist, it is also
// written there so the project works out of the box.
func copyEmbedDir(src, dst string, initLive bool) error {
	if err := copyFS(src, dst); err != nil {
		return err
	}
	if initLive {
		if _, err := os.Stat(src); os.IsNotExist(err) {
			if err := copyFS(src, src); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFS(src, dst string) error {
	return fs.WalkDir(gokesh.FS, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := gokesh.FS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return err
		}
		slog.Info("wrote", "file", target)
		return nil
	})
}
