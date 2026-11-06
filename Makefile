.PHONY: build test check dist release-check release help clean

REPO ?= xsy-tech/erp-agent-skills
VERSION ?=

build:
	go build -trimpath -buildvcs=false -o bin/erp-skill ./cmd/erp-skill

test:
	go test ./...

check:
	go test ./...
	go vet ./...

dist:
	./scripts/dist.sh

release-check:
	$(MAKE) check
	$(MAKE) dist
	bash ./scripts/verify-dist.sh

release:
	@RELEASE_REPO="$(REPO)" RELEASE_VERSION="$(VERSION)" bash ./scripts/release.sh

help:
	@printf '%s\n' \
	  'make build          Build the local CLI' \
	  'make check          Run tests and go vet' \
	  'make dist           Build archives for all five platforms' \
	  'make release-check  Check, package, and verify checksums locally' \
	  'make release        Build and publish a GitHub Release for the source version' \
	  '                    Optional: VERSION=0.1.1 REPO=xsy-tech/erp-agent-skills' \
	  'make clean          Remove bin/ and dist/'

clean:
	rm -rf bin dist
