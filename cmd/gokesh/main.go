package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/vinckr/gokesh/internal/build"
)

// version is injected at build time via -ldflags "-X main.version=vX.Y.Z"
var version = "dev"

const usage = `Usage: gokesh <command> [args]

Commands:
  init                 Set up a new project (templates, styles, config)
  build                Build all pages in markdown/ recursively
  build page <name>    Build a single page from markdown/<name>.md
  build dir <name>     Build all pages in markdown/<name>/
  new <name>           Create a new markdown file with pre-filled frontmatter
  clean                Delete the output directory
  serve                Watch for changes + serve with live reload (recommended for development)
  watch                Watch for changes and rebuild (no server)
  dev                  Serve the output directory on http://localhost:8000 (no watch)

Flags:
  --version, -v        Print version and exit
  --help,    -h        Print this help and exit
`

const templatesDir = "./templates/"

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--version", "-v":
		fmt.Println("gokesh " + version)
		return
	case "--help", "-h":
		fmt.Print(usage)
		return
	}

	if os.Args[1] == "init" {
		if err := runInit(); err != nil {
			slog.Error("init failed", "error", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := build.LoadConfig("./gokesh.toml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	mdDir := "./" + cfg.MarkdownDir + "/"
	outDir := "./" + cfg.OutputDir + "/"

	if os.Args[1] == "new" {
		if len(os.Args) < 3 {
			fmt.Println("usage: gokesh new <name>")
			os.Exit(1)
		}
		if err := runNew(os.Args[2], mdDir); err != nil {
			slog.Error("new failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if os.Args[1] == "clean" {
		if err := runClean(outDir); err != nil {
			slog.Error("clean failed", "error", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "build":
		if err := build.CopyStyles("./styles/", outDir); err != nil {
			slog.Error("failed to copy styles", "error", err)
			os.Exit(1)
		}
		if err := build.CopyStatic("./static/", outDir); err != nil {
			slog.Error("failed to copy static files", "error", err)
			os.Exit(1)
		}
		if len(os.Args) == 2 {
			if err := build.BuildAll(mdDir, outDir, cfg, templatesDir); err != nil {
				slog.Error("build failed", "error", err)
				os.Exit(1)
			}
		} else {
			if len(os.Args) < 4 {
				fmt.Print(usage)
				os.Exit(1)
			}
			switch os.Args[2] {
			case "page":
				name := os.Args[3]
				if err := build.BuildPage(name+".md", mdDir, outDir, cfg, templatesDir); err != nil {
					slog.Error("could not build page", "name", name, "error", err)
					os.Exit(1)
				}
			case "dir":
				dir := os.Args[3]
				if err := build.BuildPages(mdDir+dir+"/", outDir+dir+"/", cfg, templatesDir); err != nil {
					slog.Error("could not build directory", "dir", dir, "error", err)
					os.Exit(1)
				}
			default:
				fmt.Printf("unknown build type %q — use 'page' or 'dir'\n", os.Args[2])
				os.Exit(1)
			}
		}
		pages := build.CollectPages(mdDir, cfg.BaseURL)
		if err := build.GenerateRSS(pages, outDir, cfg.SiteTitle, cfg.BaseURL); err != nil {
			slog.Error("failed to generate RSS feed", "error", err)
			os.Exit(1)
		}
		if err := build.GenerateSitemap(outDir, cfg.BaseURL); err != nil {
			slog.Error("failed to generate sitemap", "error", err)
			os.Exit(1)
		}

	case "watch":
		if err := build.Watch(outDir, "./gokesh.toml", templatesDir); err != nil {
			slog.Error("watch failed", "error", err)
			os.Exit(1)
		}

	case "serve":
		if err := runServe(outDir, "./gokesh.toml", templatesDir, ":8000"); err != nil {
			slog.Error("serve failed", "error", err)
			os.Exit(1)
		}

	case "dev":
		addr := ":8000"
		slog.Info("serving", "addr", "http://localhost"+addr)
		http.Handle("/", http.FileServer(http.Dir(outDir)))
		if err := http.ListenAndServe(addr, nil); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("unknown command %q\n\n", os.Args[1])
		fmt.Print(usage)
		os.Exit(1)
	}
}
