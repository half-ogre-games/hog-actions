.PHONY: build clean test help

# Default target
help:
	@echo "Available targets:"
	@echo "  build      - Build all actions"
	@echo "  clean      - Remove all built binaries"
	@echo "  test       - Run tests for all actions"
	@echo "  help       - Show this help message"

# Build all actions
build: build-create-issue build-find-issue build-close-issue build-comment-issue

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

# Clean all built binaries
clean:
	@echo "Cleaning built binaries..."
	rm -f create-issue/create-issue
	rm -f find-issue/find-issue
	rm -f close-issue/close-issue
	rm -f comment-issue/comment-issue

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
	@echo "All tests completed successfully!"