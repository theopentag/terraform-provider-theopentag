default: build

build:
	go build -o terraform-provider-theopentag .

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/theopentag/theopentag/0.0.4/$(shell go env GOOS)_$(shell go env GOARCH)
	cp terraform-provider-theopentag ~/.terraform.d/plugins/registry.terraform.io/theopentag/theopentag/0.0.1/$(shell go env GOOS)_$(shell go env GOARCH)/

test:
	go test ./...

lint:
	golangci-lint run ./...

.PHONY: default build install test lint
