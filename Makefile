build:
	go build -o "./out/xns-informer-gen" ./cmd/xns-informer-gen

test:
	go vet ./...
	go test -v ./...

update-codegen:
	./hack/update-codegen.sh
