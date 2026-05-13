.PHONY: build test lint lint-fix lint-go lint-ts lint-py clean build-all run

BINARY_NAME=envguard
BUILD_DIR=bin
CMD_DIR=cmd/envguard

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

test:
	go test -v -race -coverprofile=coverage.out $$(go list ./... | grep -v node_modules)
	go tool cover -func=coverage.out

lint: lint-go lint-ts lint-py

lint-fix: lint-go-fix lint-ts-fix lint-py-fix

lint-go:
	golangci-lint run ./...

lint-go-fix:
	golangci-lint run ./... --fix

lint-ts:
	cd packages/node && npm run lint

lint-ts-fix:
	cd packages/node && npm run lint:fix

lint-py:
	cd packages/python && python3 -m ruff check envguard/ tests/
	cd packages/python && python3 -m ruff format --check envguard/ tests/

lint-py-fix:
	cd packages/python && python3 -m ruff check --fix envguard/ tests/
	cd packages/python && python3 -m ruff format envguard/ tests/

clean:
	rm -rf $(BUILD_DIR) coverage.out

build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)
