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

export CGO_ENABLED:=0

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}

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
