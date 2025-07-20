.PHONY: build test run clean

# Build the binary
build:
	go build -o close-issue main.go

# Run tests
test:
	go test -v ./...

# Run locally (example)
run: build
	./close-issue $(ARGS)

# Clean build artifacts
clean:
	rm -f close-issue

# Example usage target
example:
	@echo "Example: make run ARGS='half-ogre-games/rpgish-claude 123 \"Closing comment\" \"completed\" \"ghp_token\"'"