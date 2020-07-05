.PHONY: test version

VERSION := $(shell go run ./cmd/binq/*.go version | awk '{print $$2}')

test:
	go test -v ./...

version:
	git commit -m $(VERSION)
	git tag -a v$(VERSION) -m $(VERSION)
	git push origin v$(VERSION)
	git push origin master
