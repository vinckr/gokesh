package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/vinckr/gokesh/internal/build"
)

const usage = `Usage: gokesh <command> [args]

Commands:
  build                Build all pages in markdown/ recursively
  build page <name>    Build a single page from markdown/<name>.md
  build dir <name>     Build all pages in markdown/<name>/
  watch                Watch for changes and rebuild automatically
  dev                  Serve public/ on http://localhost:8000
`

var defaultTemplates = []string{
	"./templates/page.tmpl",
	"./templates/header.tmpl",
	"./templates/footer.tmpl",
	"./templates/body.tmpl",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	cfg, err := build.LoadConfig("./gokesh.toml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		if err := build.CopyStyles("./styles/", "./public/"); err != nil {
			slog.Error("failed to copy styles", "error", err)
			os.Exit(1)
		}
		if len(os.Args) == 2 {
			if err := build.BuildAll("./markdown/", "./public/", cfg, defaultTemplates...); err != nil {
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
				if err := build.BuildPage(name+".md", "./markdown/", "./public/", cfg, defaultTemplates...); err != nil {
					slog.Error("could not build page", "name", name, "error", err)
					os.Exit(1)
				}
			case "dir":
				dir := os.Args[3]
				if err := build.BuildPages("./markdown/"+dir+"/", "./public/", cfg, defaultTemplates...); err != nil {
					slog.Error("could not build directory", "dir", dir, "error", err)
					os.Exit(1)
				}
			default:
				fmt.Printf("unknown build type %q — use 'page' or 'dir'\n", os.Args[2])
				os.Exit(1)
			}
		}
		if err := build.GenerateSitemap("./public/", cfg.BaseURL); err != nil {
			slog.Error("failed to generate sitemap", "error", err)
			os.Exit(1)
		}

	case "watch":
		if err := build.Watch("./public/", cfg, defaultTemplates...); err != nil {
			slog.Error("watch failed", "error", err)
			os.Exit(1)
		}

	case "dev":
		addr := ":8000"
		slog.Info("serving", "addr", "http://localhost"+addr)
		http.Handle("/", http.FileServer(http.Dir("./public")))
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
