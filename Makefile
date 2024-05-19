PROJECT_NAME := "taskgroup"
PKG := "github.com/mlee-msl/taskgroup"
PKG_LIST := $(shell go mod tidy && go list ${PKG}/...)

.DEFAULT_GOAL := default
.PHONY: all

all: fmt vet test race build

dep: ## Get dependencies
	@echo "go dep..."
	@go mod tidy

fmt: dep ## Format code
	@echo "go fmt..."
	@go fmt $(PKG_LIST)

vet: dep ## Vet check
	@echo "go vet..."
	@go vet -all $(PKG_LIST)

test: dep ## Run unittests
	@echo "go test..."
	@go test -gcflags=all=-l -short -v -count=1 ${PKG_LIST}

race: dep ## Run data race detector
	@echo "go test race..."
	@go test -gcflags=all=-l -race -short -v -count=1 ${PKG_LIST}

build: dep fmt ## Build
	@echo "go build..."
	@go build -v -buildmode=default -o ${PROJECT_NAME} ./...
#	@chmod +x bin/${PROJECT_NAME}

run: dep fmt ## Run
	@echo "go run..."
	@go run ./...

help: 
	@echo "make -k <all>|<build>|<test>|..."

default: help