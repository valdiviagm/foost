# Go command
GOCMD=go

# Default action: run the server
all:
	$(GOCMD) run main.go

.PHONY: all

setup:
	$(GOCMD) get github.com/onsi/ginkgo/ginkgo@latest
	$(GOCMD) get github.com/onsi/gomega/...@latest

