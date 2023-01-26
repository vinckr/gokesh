help:  # show all available Make commands
	@cat Makefile | grep '^[^ ]*:' | grep -v '^\.bin/' | grep -v '^node_modules' | grep -v '.SILENT:' | grep -v help | sed 's/:.*#/#/' | column -s "#" -t

test: # test building HTML files
	echo "Building HTML files"
	go run cmd/build/main.go page index
	go run cmd/build/main.go dir blog

dev: # run a local server to preview the site
	make test
	@echo "\nPreview running at http://localhost:8000"
	go run cmd/dev/main.go
