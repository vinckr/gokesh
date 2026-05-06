package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var smokeBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gokesh-smoke-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "smoke: mkdirtemp: %v\n", err)
		os.Exit(1)
	}
	name := "gokesh"
	if runtime.GOOS == "windows" {
		name = "gokesh.exe"
	}
	smokeBin = filepath.Join(dir, name)
	out, err := exec.Command("go", "build", "-o", smokeBin, ".").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "smoke: go build: %v\n%s\n", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func smokeRun(t *testing.T, dir string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command(smokeBin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	stdout = string(out)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout, exitErr.ExitCode()
		}
		t.Fatalf("exec error: %v", err)
	}
	return stdout, 0
}

func TestSmokeVersion(t *testing.T) {
	out, code := smokeRun(t, t.TempDir(), "--version")
	if code != 0 {
		t.Fatalf("--version exited %d, want 0; output: %s", code, out)
	}
	if !strings.Contains(out, "gokesh") {
		t.Errorf("--version output missing 'gokesh': %s", out)
	}
}

func TestSmokeHelp(t *testing.T) {
	out, code := smokeRun(t, t.TempDir(), "--help")
	if code != 0 {
		t.Fatalf("--help exited %d, want 0; output: %s", code, out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Errorf("--help output missing 'Commands:': %s", out)
	}
}

func TestSmokeInit(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command(smokeBin, "init")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("n\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init failed: %v\noutput: %s", err, out)
	}
	for _, want := range []string{"gokesh.toml", "templates", "styles"} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("init: %s not created: %v", want, err)
		}
	}
}

func TestSmokeBuildAfterInit(t *testing.T) {
	dir := t.TempDir()

	cmd := exec.Command(smokeBin, "init")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("n\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\noutput: %s", err, out)
	}

	out, code := smokeRun(t, dir, "build")
	if code != 0 {
		t.Fatalf("build exited %d; output: %s", code, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "public", "index.html")); err != nil {
		t.Errorf("build: public/index.html not created: %v", err)
	}
}

func TestSmokeNew(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gokesh.toml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "markdown"), 0755); err != nil {
		t.Fatal(err)
	}

	out, code := smokeRun(t, dir, "new", "my-post")
	if code != 0 {
		t.Fatalf("new exited %d; output: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "markdown", "my-post.md"))
	if err != nil {
		t.Fatalf("markdown/my-post.md not created: %v", err)
	}
	if !strings.Contains(string(data), `title: "my-post"`) {
		t.Errorf("my-post.md missing title frontmatter:\n%s", data)
	}
}

func TestSmokeClean(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gokesh.toml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "public"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "public", "index.html"), []byte("<html>"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(smokeBin, "clean")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("y\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clean failed: %v\noutput: %s", err, out)
	}

	if _, err := os.Stat(filepath.Join(dir, "public")); !os.IsNotExist(err) {
		t.Error("clean: public/ still exists after clean")
	}
}
