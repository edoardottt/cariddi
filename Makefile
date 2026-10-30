REPO := github.com/edoardottt/cariddi/v2
BIN := cariddi
BUILD_DIR := ./cmd/$(BIN)
INSTALL_PATH := /usr/bin/$(BIN)

.PHONY: tidy remod update lint linux unlinux test

# Update only the direct dependencies (see update).
tidy:
	@echo "Tidying modules..."
	@go get $$(go list -m -f '{{if not .Main}}{{if not .Indirect}}{{.Path}}@latest{{end}}{{end}}' all)
	@go mod tidy -v

remod:
	@echo "Reinitializing modules..."
	@rm -f go.mod go.sum
	@go mod init $(REPO)
	@go get
	@echo "Done."

# Pull first, update only the direct dependencies (their dependencies move
# only as far as they require: go get -u can pick an incompatible version),
# then install: if the build fails, the installed binary is kept.
update:
	@git pull
	@echo "Updating modules..."
	@go get $$(go list -m -f '{{if not .Main}}{{if not .Indirect}}{{.Path}}@latest{{end}}{{end}}' all)
	@go mod tidy -v
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