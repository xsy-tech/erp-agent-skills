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
	  'make build          构建本机命令行工具' \
	  'make check          运行测试和 go vet' \
	  'make dist           生成全部五个平台的安装包' \
	  'make release-check  本地检查、打包并验证 SHA-256' \
	  'make release        按源码版本构建并发布 GitHub Release' \
	  '                    可选：VERSION=0.1.1 REPO=xsy-tech/erp-agent-skills' \
	  'make clean          清理 bin/ 和 dist/'

clean:
	rm -rf bin dist
