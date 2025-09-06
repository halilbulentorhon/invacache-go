.PHONY: test fieldalignment clean

# Run tests
test:
	go test -race ./...

# Check struct field alignment (excluding mocks and test files)
fieldalignment:
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@v0.14.0
	@for dir in backend config constant pkg invalidation; do \
		fieldalignment ./$$dir 2>/dev/null || true; \
	done | grep -v "_test.go:" || true

# Clean everything
clean:
	go clean
	rm -f coverage.out
