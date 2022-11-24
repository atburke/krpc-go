.PHONY: gen fmt test gen-clean

gen:
	go generate ./...

gen-clean:
	rm lib/service/*/*.gen.go

fmt:
	gofmt -w .

test:
	go test ./...