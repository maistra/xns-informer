clean:
	rm -rf vendor/ out/

deps:
	go mod download
	go mod vendor
	go mod tidy -go=1.19

build: deps
	go build -o "./out/xns-informer-gen" ./cmd/xns-informer-gen

test: build
	go vet ./...
	go test -race -v ./...

gen: build
	./hack/update-codegen.sh

gen-check: gen check-clean-repo

check-clean-repo:
	@./hack/check_clean_repo.sh

.DEFAULT_GOAL:=all
.PHONY: all
all: clean build test gen gen-check check-clean-repo
