.PHONY: build test run clean

# Build the binary
build:
	go build -o comment-issue main.go

# Run tests
test:
	go test -v ./...

# Run locally (example)
run: build
	./comment-issue $(ARGS)

# Clean build artifacts
clean:
	rm -f comment-issue

# Example usage target
example:
	@echo "Example: make run ARGS='half-ogre-games/rpgish-claude 123 \"Test comment\" \"ghp_token\"'"