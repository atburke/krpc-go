.PHONY: build gen fmt test
.DEFAULT: build

build:
	go build -o build/marshal cmd/marshal.go

gen:
	go generate ./...

fmt:
	gofmt -w .

test:
	go test ./...