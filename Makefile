.PHONY: gen fmt test integration gen-clean

gen:
	go generate ./...

gen-clean:
	rm ./*/*.gen.go

fmt:
	gofmt -w .

test:
	go test . ./lib/...

integration:
	go test ./integration