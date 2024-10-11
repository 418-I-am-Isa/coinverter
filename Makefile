# Variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=coinverter
BINARY_PATH=./$(BINARY_NAME)

# Targets
all: build

build:
	$(GOBUILD) -o $(BINARY_PATH) ./...

run:
	$(BINARY_PATH)

clean:
	$(GOCLEAN)
	rm -rf $(BINARY_PATH)

test:
	$(GOTEST) ./... -v

.PHONY: all build run clean test
