build:
	go build -o build/webtransport

fmt:
	gofmt -w .

.PHONY: build
