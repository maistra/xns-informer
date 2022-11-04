FINDFILES=find . \( -path ./.git -o -path ./.github -o -path ./tmp -o -path ./vendor \) -prune -o -type f
XARGS = xargs -0 -r

clean:
	rm -rf out/

deps:
	go mod tidy -go=1.19
	go mod download

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

################################################################################
# linting
################################################################################
.PHONY: lint-scripts
lint-scripts:
	@${FINDFILES} -name '*.sh' -print0 | ${XARGS} shellcheck

.PHONY: lint-go
lint-go:
	${FINDFILES} -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' -o -name 'zz_generated.*.go' \) \) -print0 | ${XARGS} ./hack/lint_go.sh

.PHONY: lint
lint: lint-scripts lint-go

.DEFAULT_GOAL:=all
.PHONY: all
all: clean build test lint gen gen-check check-clean-repo
