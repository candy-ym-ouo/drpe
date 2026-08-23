.PHONY: build test demo size
build:
	go build -o drpe ./cmd/drpe
test:
	go test ./...
demo:
	go run ./cmd/drpe -demo
size:
	@files=$$(find . -name '*.go' ! -name '*_test.go' | wc -l); lines=$$(find . -name '*.go' ! -name '*_test.go' -print0 | xargs -0 cat | wc -l); echo "go files=$$files lines=$$lines"; test $$files -gt 20 -a $$files -lt 25; test $$lines -gt 2000 -a $$lines -lt 2200
