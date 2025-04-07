REPO := github.com/edoardottt/cariddi
BIN := cariddi
BUILD_DIR := ./cmd/$(BIN)
INSTALL_PATH := /usr/bin/$(BIN)

.PHONY: tidy remod update lint linux unlinux test

tidy:
	@echo "Tidying modules..."
	@go get -u ./...
	@go mod tidy -v

remod:
	@echo "Reinitializing modules..."
	@rm -f go.mod go.sum
	@go mod init $(REPO)
	@go get
	@echo "Done."

update: unlinux
	@echo "Updating modules..."
	@go get -u ./...
	@go mod tidy -v
	@git pull
	@$(MAKE) linux
	@echo "Done."

lint:
	@echo "Running linter..."
	@golangci-lint run

linux:
	@echo "Building for Linux..."
	@go build -o $(BIN) $(BUILD_DIR)
	@sudo mv $(BIN) $(INSTALL_PATH)
	@echo "Installed at $(INSTALL_PATH)."

linux-arm64:
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(BIN)-arm64 $(BUILD_DIR)
	@echo "Output: $(BIN)-arm64"

windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(BIN).exe $(BUILD_DIR)
	@echo "Output: $(BIN).exe"

unlinux:
	@echo "Removing Linux binary..."
	@sudo rm -f $(INSTALL_PATH)
	@echo "Done."

test:
	@echo "Running tests..."
	@go test -race ./...
	@echo "Done."

clean:
	@echo "Cleaning up build artifacts..."
	@rm -f $(BIN) $(BIN).exe $(BIN)-arm64
	@echo "Cleaned."