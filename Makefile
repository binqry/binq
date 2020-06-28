.PHONY: version

VERSION := $(shell go run ./cmd/binq/*.go -V | awk '{print $$2}')

version:
	git commit -m $(VERSION)
	git tag -a v$(VERSION) -m $(VERSION)
	git push origin v$(VERSION)
	git push origin master
