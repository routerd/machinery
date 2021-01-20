# routerd
# Copyright (C) 2020  The routerd Authors
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

SHELL=/bin/bash
.SHELLFLAGS=-euo pipefail -c

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_BIN = $(TMP)/bin
TMP_VERSIONS := $(TMP)/versions

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}

PROTOC_VERSION:=v3.13.0
PROTOC_GENGO_VERSION:=v1.25.0
PROTOC_GENGO_GRPC_VERSION:=v1.0.1
PROTOC_GEN_GRPCGATEWAY_VERSION:=v2.1.0
PROTOC_GEN_OPENAPIV2_VERSION:=${PROTOC_GEN_GRPCGATEWAY_VERSION}

export CGO_ENABLED:=0
export GO111MODULE:=on
export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)

# -------------------
# Testing and Linting
# -------------------

test:
	CGO_ENABLED=1 go test -race -v ./...
.PHONY: test

fmt:
	go fmt ./...
.PHONY: fmt

vet:
	go vet ./...
.PHONY: vet

tidy:
	go mod tidy
.PHONY: tidy

verify-boilerplate:
	@go run hack/boilerplate/boilerplate.go \
		-boilerplate-dir hack/boilerplate/ \
		-verbose
.PHONY: verify-boilerplate

pre-commit-install:
	@echo "installing pre-commit hooks using https://pre-commit.com/"
	@pre-commit install
.PHONY: pre-commit-install

# ------------
# Dependencies
# ------------


PROTOC := $(TMP_VERSIONS)/protoc/$(PROTOC_VERSION)
$(PROTOC):
	$(eval PROTOC_TMP := $(shell mktemp -d))
	@mkdir -p $(TMP_BIN)
	@ cd $(PROTOC_TMP); \
		curl -L --fail \
	 	https://github.com/protocolbuffers/protobuf/releases/download/$(PROTOC_VERSION)/protoc-$(shell echo -n $(PROTOC_VERSION) | tail -c +2)-linux-$(UNAME_ARCH).zip \
		-o protoc.zip && \
		unzip -qq protoc.zip
	@mv $(PROTOC_TMP)/bin/protoc $(TMP_BIN)
	@mv $(PROTOC_TMP)/include $(TMP)
	@rm -rf $(PROTOC_TMP)
	@rm -rf $(dir $(PROTOC))
	@mkdir -p $(dir $(PROTOC))
	@touch $(PROTOC)

PROTOC_GENGO_ := $(TMP_VERSIONS)/protoc-gen-go/$(PROTOC_GENGO_VERSION)
$(PROTOC_GENGO_):
	$(eval PROTOC_GENGO__TMP := $(shell mktemp -d))
	@cd $(PROTOC_GENGO__TMP); go get google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GENGO_VERSION)
	@rm -rf $(PROTOC_GENGO__TMP)
	@rm -rf $(dir $(PROTOC_GENGO_))
	@mkdir -p $(dir $(PROTOC_GENGO_))
	@touch $(PROTOC_GENGO_)

PROTOC_GENGO_GRPC := $(TMP_VERSIONS)/protoc-gen-go-grpc/$(PROTOC_GENGO_GRPC_VERSION)
$(PROTOC_GENGO_GRPC):
	$(eval PROTOC_GENGO_GRPC_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GENGO_GRPC_TMP); go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GENGO_GRPC_VERSION)
	@rm -rf $(PROTOC_GENGO_GRPC_TMP)
	@rm -rf $(dir $(PROTOC_GENGO_GRPC))
	@mkdir -p $(dir $(PROTOC_GENGO_GRPC))
	@touch $(PROTOC_GENGO_GRPC)

PROTOC_GEN_GRPCGATEWAY := $(TMP_VERSIONS)/protoc-gen-go-grpc-gateway/$(PROTOC_GEN_GRPCGATEWAY_VERSION)
$(PROTOC_GEN_GRPCGATEWAY):
	$(eval PROTOC_GEN_GRPCGATEWAY_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GEN_GRPCGATEWAY_TMP); go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPCGATEWAY_VERSION)
	@rm -rf $(PROTOC_GEN_GRPCGATEWAY_TMP)
	@rm -rf $(dir $(PROTOC_GEN_GRPCGATEWAY))
	@mkdir -p $(dir $(PROTOC_GEN_GRPCGATEWAY))
	@touch $(PROTOC_GEN_GRPCGATEWAY)

PROTOC_GEN_OPENAPIV2 := $(TMP_VERSIONS)/protoc-gen-openapi-v2/$(PROTOC_GEN_OPENAPIV2_VERSION)
$(PROTOC_GEN_OPENAPIV2):
	$(eval PROTOC_GEN_OPENAPIV2_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GEN_OPENAPIV2_TMP); go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(PROTOC_GEN_OPENAPIV2_VERSION)
	@rm -rf $(PROTOC_GEN_OPENAPIV2_TMP)
	@rm -rf $(dir $(PROTOC_GEN_OPENAPIV2))
	@mkdir -p $(dir $(PROTOC_GEN_OPENAPIV2))
	@touch $(PROTOC_GEN_OPENAPIV2)

PROTO_GRPC_GATEWAY_INCLUDES := $(TMP_VERSIONS)/proto-grpc-gateway-includes/$(PROTOC_GEN_GRPCGATEWAY_VERSION)
$(PROTO_GRPC_GATEWAY_INCLUDES):
	$(eval PROTO_GRPC_GATEWAY_INCLUDES_TMP := $(shell mktemp -d))
	@cd $(PROTO_GRPC_GATEWAY_INCLUDES_TMP); git clone https://github.com/grpc-ecosystem/grpc-gateway --depth=1 --branch=$(PROTOC_GEN_GRPCGATEWAY_VERSION) .
	@cp -a $(PROTO_GRPC_GATEWAY_INCLUDES_TMP)/third_party/googleapis/google/* $(TMP)/include/google
	@rm -rf $(PROTO_GRPC_GATEWAY_INCLUDES_TMP)
	@rm -rf $(dir $(PROTO_GRPC_GATEWAY_INCLUDES))
	@mkdir -p $(dir $(PROTO_GRPC_GATEWAY_INCLUDES))
	@touch $(PROTO_GRPC_GATEWAY_INCLUDES)

# ----------
# Generators
# ----------

%.proto: FORCE $(PROTOC) $(PROTOCGENGO) $(PROTOCGENGOGRPC) $(PROTOC_GEN_GRPCGATEWAY) $(PROTOC_GEN_OPENAPIV2) $(PROTO_GRPC_GATEWAY_INCLUDES)
	$(eval PROTO_DIR := $(shell dirname $@))
	@ echo generating $@
	@protoc \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_DIR) \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true \
		-I$(TMP)/include \
		-Iapi/v1 \
		-I$(PROTO_DIR) $@

	$(eval PROTO_FILEBASE := $(PROTO_DIR)/$(shell basename $@ .proto))
	@# add boilerplate to _grpc.pb.go files
	@test -s $(PROTO_FILEBASE)_grpc.pb.go && \
		cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/2020/ > $(PROTO_FILEBASE)_grpc.pb.go.tmp && \
		cat $(PROTO_FILEBASE)_grpc.pb.go >> $(PROTO_FILEBASE)_grpc.pb.go.tmp && \
		mv $(PROTO_FILEBASE)_grpc.pb.go.tmp $(PROTO_FILEBASE)_grpc.pb.go && \
		goimports -local routerd.net -w $(PROTO_FILEBASE)_grpc.pb.go; true

	@# add boilerplate to .pb.gw.go files
	@test -s $(PROTO_FILEBASE).pb.gw.go && \
		cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/2020/ > $(PROTO_FILEBASE).pb.gw.go.tmp && \
		cat $(PROTO_FILEBASE).pb.gw.go >> $(PROTO_FILEBASE).pb.gw.go.tmp && \
		mv $(PROTO_FILEBASE).pb.gw.go.tmp $(PROTO_FILEBASE).pb.gw.go && \
		goimports -local routerd.net -w $(PROTO_FILEBASE).pb.gw.go; true

	@goimports -local routerd.net -w $(PROTO_FILEBASE).pb.go

.PHONY: proto
proto: api/**/*.proto testdata/**/*.proto

FORCE:
