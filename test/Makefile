GOPATH ?= ${GOPATH}
BINDIR ?= ${GOPATH}/bin

.PHONY: all update-vendor dep clean install test

VERSION=$(shell git describe --tags --always --dirty)

test:
	go test -v github.com/opencontainers/distribution-spec/test/dist/...

update-vendor: | dep
	dep ensure
dep:
	@which dep || go get -u github.com/golang/dep/cmd/dep

clean:
	rm -f $(BIN)

install:
	install $(BIN) "$(BINDIR)"
