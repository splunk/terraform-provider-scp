SHELL = /bin/bash

OPENAPI_JSON_URL=https://admin.splunk.com/service/info/specs/v2/openapi.json

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OS=$(shell uname | tr A-Z a-z)

ARCH=$(shell uname -m)
ifeq ($(ARCH), x86_64)
ARCH=amd64
endif

####################################
#	Building binary
####################################
vendor:
	go mod vendor

generate: vendor oapi-codegen
	# generate the openapi related code
	$(OAPI_CODEGEN) --generate types,client --package v2 $(OPENAPI_JSON_URL) > acs/v2/api.gen.go
	# generate the mocks for api.ClientInterface
	$(MOCKERY) --dir acs/v2 --name ClientInterface --output acs/v2/mocks

default: build

fmt:
	go fmt ./...
	@terraform fmt -recursive

#local-build: clean generate
#	go build -o terraform.local/local/splunkcloud/1.0.0/$(OS)_$(ARCH)/terraform-provider-splunkcloud .
#	rm -rf ~/.terraform.d/plugins/terraform.local
#	mv terraform.local/ ~/.terraform.d/plugins

build:
	go build -o terraform-provider-splunkcloud .

#clean:
#	rm -f .terraform.lock.hcl
#	rm -rf .terraform/providers

###################################
#	Testing commands
###################################

#run unit tests
test: go-junit-report
	go test -short -covermode=atomic -coverprofile=./coverage.txt ./... -v 2>&1 | tee ./test.txt && ./scripts/exclude-from-unit-test.sh
	cat test.txt | $(GO_JUNIT_REPORT) > ./report.xml
	go tool cover -func=./coverage.txt

#run acceptance tests
testacc:
	TF_ACC=1 go test -run "^TestAcc" ./... -v

###################################
#	Install dependency
###################################

# find or download oapi-codegen
# See https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies, go doesn't use `go get` to install packages anymore.
oapi-codegen:
ifneq (0, $(shell command -v oapi-codegen ; echo $$?))
	@ go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.6.0
OAPI_CODEGEN=$(GOBIN)/oapi-codegen
else
OAPI_CODEGEN=$(shell which oapi-codegen)
endif

# find or download go-junit-report
go-junit-report:
ifneq (0, $(shell command -v go-junit-report ; echo $$?))
	@ go install github.com/jstemmer/go-junit-report@latest
GO_JUNIT_REPORT=$(GOBIN)/go-junit-report
else
GO_JUNIT_REPORT=$(shell which go-junit-report)
endif

# find or download mockery
mockery:
ifneq (0, $(shell command -v mockery ; echo $$?))
	@ go install github.com/vektra/mockery/v2/...@v2.9.5
MOCKERY=$(GOBIN)/mockery
else
MOCKERY=$(shell which mockery)
endif

## Run acceptance tests
#.PHONY: testacc
