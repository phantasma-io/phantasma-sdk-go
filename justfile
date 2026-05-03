[private]
just:
    just -l

[group('doc')]
guide:
    cat README.md | less

[group('format')]
f:
    find . -type f -name '*.go' -not -path './.git/*' -not -path './pkg/util/bigint_test_data.go' -exec gofmt -w {} +

format: f

[group('build')]
build:
    go build ./...

build-verbose:
    go build -v ./...

clean:
    go clean -cache
    go clean -testcache

[group('test')]
test:
    go test ./...

test-clean:
    go clean -testcache
    go test ./...

[group('run')]
run-example:
    go run ./examples
