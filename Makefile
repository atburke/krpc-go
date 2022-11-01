.PHONY: build gen
.DEFAULT: build

build:
	go build -o build/marshal cmd/marshal.go

gen:
	go generate ./...