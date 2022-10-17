.PHONY: build
build:
	go build -v .

.PHONY: test
test:
	go test -v ./...

dep-upgrade:
	go get -u
	go mod tidy
