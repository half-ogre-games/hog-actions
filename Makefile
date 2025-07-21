.PHONY: build clean test help build-create-issue build-find-issue build-close-issue build-comment-issue build-get-latest-semver-tag build-get-next-semver build-tag-and-create-semver-release

# Default target
help:
	@echo "Available targets:"
	@echo "  build      - Build all actions"
	@echo "  clean      - Remove all built binaries"
	@echo "  test       - Run tests for all actions"
	@echo "  help       - Show this help message"

# Build all actions
build: build-create-issue build-find-issue build-close-issue build-comment-issue build-get-latest-semver-tag build-get-next-semver build-tag-and-create-semver-release

build-create-issue:
	@echo "Building create-issue..."
	cd create-issue && go build -o create-issue main.go

build-find-issue:
	@echo "Building find-issue..."
	cd find-issue && go build -o find-issue main.go

build-close-issue:
	@echo "Building close-issue..."
	cd close-issue && go build -o close-issue main.go

build-comment-issue:
	@echo "Building comment-issue..."
	cd comment-issue && go build -o comment-issue main.go

build-get-latest-semver-tag:
	@echo "Building get-latest-semver-tag..."
	cd get-latest-semver-tag && go build -o get-latest-semver-tag main.go

build-get-next-semver:
	@echo "Building get-next-semver..."
	cd get-next-semver && go build -o get-next-semver main.go

build-tag-and-create-semver-release:
	@echo "Building tag-and-create-semver-release..."
	cd tag-and-create-semver-release && go build -o tag-and-create-semver-release main.go

# Clean all built binaries
clean:
	@echo "Cleaning built binaries..."
	rm -f create-issue/create-issue
	rm -f find-issue/find-issue
	rm -f close-issue/close-issue
	rm -f comment-issue/comment-issue
	rm -f get-latest-semver-tag/get-latest-semver-tag
	rm -f get-next-semver/get-next-semver
	rm -f tag-and-create-semver-release/tag-and-create-semver-release

# Run tests for all actions and go-kit
test:
	@echo "Running tests..."
	@echo "Testing go-kit..."
	@cd internal/go-kit && go test -v ./...
	@echo "Testing create-issue..."
	@cd create-issue && go test -v ./...
	@echo "Testing find-issue..."
	@cd find-issue && go test -v ./...
	@echo "Testing close-issue..."
	@cd close-issue && go test -v ./...
	@echo "Testing comment-issue..."
	@cd comment-issue && go test -v ./...
	@echo "Testing get-latest-semver-tag..."
	@cd get-latest-semver-tag && go test -v ./...
	@echo "Testing get-next-semver..."
	@cd get-next-semver && go test -v ./...
	@echo "Testing semveractions..."
	@cd internal/semveractions && go test -v ./...
	@echo "Testing tag-and-create-semver-release..."
	@cd tag-and-create-semver-release && go test -v ./...
	@echo "All tests completed successfully!"