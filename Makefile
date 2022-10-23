.PHONY: build gen
.DEFAULT: build

build:
	go build -o build/listservices cmd/listservices.go

gen:
	go generate ./...