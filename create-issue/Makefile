.PHONY: build test run clean

# Build the binary
build:
	go build -o create-issue main.go

# Run tests
test:
	go test -v ./...

# Run locally (example)
run: build
	./create-issue

# Clean build artifacts
clean:
	rm -f create-issue

# Example usage target
example:
	@echo "Set environment variables then run:"
	@echo "export GITHUB_REPOSITORY='half-ogre-games/rpgish-claude'"
	@echo "export INPUT_ISSUE_TITLE='Test Issue'"
	@echo "export INPUT_ISSUE_BODY='Issue body'"
	@echo "export INPUT_ISSUE_LABEL='bug'"
	@echo "export INPUT_ADDITIONAL_LABELS='priority-high'"
	@echo "export INPUT_GITHUB_TOKEN='ghp_token'"
	@echo "make run"