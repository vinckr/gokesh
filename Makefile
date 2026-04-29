.PHONY: help test vet build dev install update-golden release

help: # show all available Make commands
	@cat Makefile | grep '^[^ ]*:' | grep -v '^\.bin/' | grep -v '^node_modules' | grep -v '.SILENT:' | grep -v help | sed 's/:.*#/#/' | column -s "#" -t

test: # run tests
	go test ./...

vet: # run go vet
	go vet ./...

build: # build the gokesh binary
	go build -o bin/gokesh ./cmd/gokesh

dev: # build example pages and start preview server at localhost:8000
	go run ./cmd/gokesh build page index
	go run ./cmd/gokesh build dir blog
	@echo "\nPreview running at http://localhost:8000"
	go run ./cmd/gokesh dev

install: # install gokesh binary to $GOPATH/bin
	go install ./cmd/gokesh

update-golden: # update golden test files
	go test ./internal/build/ -update

release: test vet # tag and push a new release (usage: make release VERSION=v0.1.0)
	@test -n "$(VERSION)" || (echo "VERSION is required — usage: make release VERSION=v0.1.0" && exit 1)
	@echo "Releasing $(VERSION)..."
	git tag $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Release $(VERSION) pushed — GitHub Actions will build and publish the binaries."
