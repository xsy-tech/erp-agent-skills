.PHONY: build test check dist clean

build:
	go build -trimpath -buildvcs=false -o bin/erp-skill ./cmd/erp-skill

test:
	go test ./...

check:
	go test ./...
	go vet ./...

dist:
	./scripts/dist.sh

clean:
	rm -rf bin dist
