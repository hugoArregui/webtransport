build:
	go build ./pkg/webtransport/

test: build
	go test $(TEST_FLAGS) -count=1 -race test/webtransport_test.go

fmt:
	gofmt -w .
	goimports -w .

.PHONY: build
