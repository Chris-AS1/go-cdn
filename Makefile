GO := go
BUILD := $(GO) build
CLEAN := $(GO) clean
BUILD_TEST := $(GO) test -c
RUN_TEST := $(wildcard ./*.test) 

BINARY := go-cdn

default: build

build:
	@echo "Building $(BINARY)..."
	@$(BUILD) -o $(BINARY) ./cmd/go-cdn/main.go 
clean:
	@echo "Cleaning up..."
	$(CLEAN)
	@rm -f $(BINARY)
	@for testfile in $(RUN_TEST); do \
			rm $$testfile; \
	done
test:
	@echo "Building and running tests..."
	@$(BUILD_TEST) ./tests/*
	@for testfile in *.test; do \
		./$$testfile -test.v; \
	done

.PHONY: build clean test
