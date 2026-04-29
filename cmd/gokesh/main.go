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
  build page <name>   Build a single page from markdown/<name>.md
  build dir <name>    Build all pages in markdown/<name>/
  dev                 Serve public/ on http://localhost:8000
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

	switch os.Args[1] {
	case "build":
		if len(os.Args) < 4 {
			fmt.Print(usage)
			os.Exit(1)
		}
		if err := build.CopyStyles("./styles/", "./public/"); err != nil {
			slog.Error("failed to copy styles", "error", err)
			os.Exit(1)
		}
		var err error
		switch os.Args[2] {
		case "page":
			err = build.BuildPage(os.Args[3]+".md", "./markdown/", "./public/", defaultTemplates...)
		case "dir":
			err = build.BuildPages("markdown/"+os.Args[3]+"/", "./public/", defaultTemplates...)
		default:
			fmt.Printf("unknown build type %q — use 'page' or 'dir'\n", os.Args[2])
			os.Exit(1)
		}
		if err != nil {
			slog.Error("build failed", "error", err)
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
