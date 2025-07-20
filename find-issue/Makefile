.PHONY: build test run clean

# Build the binary
build:
	go build -o find-issue main.go

# Run tests
test:
	go test -v ./...

# Run locally (example)
run: build
	./find-issue $(ARGS)

# Clean build artifacts
clean:
	rm -f find-issue

# Example usage target
example:
	@echo "Example: make run ARGS='half-ogre-games/rpgish-claude \"🚨 Terraform Drift Detected\" \"ghp_token\"'"
