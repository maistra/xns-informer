build:
	go build -o "./out/xns-informer-gen" ./cmd/xns-informer-gen

test:
	go vet ./...
	go test -race -v ./...

gen: build
	./hack/update-codegen.sh

gen-check: gen check-clean-repo

check-clean-repo:
	@./hack/check_clean_repo.sh

.PHONY: build test gen gen-check check-clean-repo
