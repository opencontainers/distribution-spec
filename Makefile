EPOCH_TEST_COMMIT	:= 91d6d8466e68f1efff7977b63ad6f48e72245e05
CURRENT_COMMIT	:= $(shell git log --format="%H" -n 1)
CONFORMANCE_VERSION ?= $(CURRENT_COMMIT)

DOCKER	?= $(shell command -v docker 2>/dev/null)
PANDOC	?= $(shell command -v pandoc 2>/dev/null)
GOLANGCILINT	?= $(shell command -v golangcli-lint 2>/dev/null)

OUTPUT_DIRNAME	?= output/
DOC_FILENAME	?= oci-distribution-spec

PANDOC_CONTAINER ?= ghcr.io/opencontainers/pandoc:2.9.2.1-9.fc34.x86_64@sha256:590c5c7aaa6e8e7a4debae7e9102c837daa0c8a76f8f5b5c9831ea5f755e3e95
ifeq "$(strip $(PANDOC))" ''
	ifneq "$(strip $(DOCKER))" ''
		PANDOC = $(DOCKER) run \
			--rm \
			-v $(shell pwd)/:/input/:ro \
			-v $(shell pwd)/$(OUTPUT_DIRNAME)/:/$(OUTPUT_DIRNAME)/ \
			-u $(shell id -u) \
			--workdir /input \
			$(PANDOC_CONTAINER)
		PANDOC_SRC := /input/
		PANDOC_DST := /
	endif
endif

GOLANGCILINT_CONTAINER ?= ghcr.io/opencontainers/golangci-lint:v1.52.1@sha256:d3d3d56f9706ebe843c1b06686c385877ba65b33f39507cdbeb22f482adce65a
ifeq "$(strip $(GOLANGCILINT))" ''
	ifneq "$(strip $(DOCKER))" ''
		GOLANGCILINT = $(DOCKER) run \
			--rm \
			-v $(shell pwd)/:/input:ro \
			-e GOCACHE=/tmp/.cache \
			-e GO111MODULE=on \
			-e GOLANGCI_LINT_CACHE=/tmp/.cache \
			--entrypoint /bin/bash \
			-u $(shell id -u) \
			--workdir /input \
			$(GOLANGCILINT_CONTAINER)
		GOLANGCILINT_SRC := /input/
		GOLANGCILINT_DST := /
	endif
endif

DOC_FILES	:= spec.md
FIGURE_FILES	:=

test: .gitvalidation

# When this is running in GitHub, it will only check the GitHub commit range
.gitvalidation:
	@command -v git-validation >/dev/null 2>/dev/null || (echo "ERROR: git-validation not found. Consider 'make install.tools' target" && false)
ifdef GITHUB_SHA
	git-validation -q -run DCO,short-subject,dangling-whitespace -range $(GITHUB_SHA)..HEAD
else
	git-validation -v -run DCO,short-subject,dangling-whitespace -range $(EPOCH_TEST_COMMIT)..HEAD
endif

docs: $(OUTPUT_DIRNAME)/$(DOC_FILENAME).pdf $(OUTPUT_DIRNAME)/$(DOC_FILENAME).html

ifeq "$(strip $(PANDOC))" ''
$(OUTPUT_DIRNAME)/$(DOC_FILENAME).pdf: $(DOC_FILES) $(FIGURE_FILES)
	$(error cannot build $@ without either pandoc or docker)
else
$(OUTPUT_DIRNAME)/$(DOC_FILENAME).pdf: $(DOC_FILES) $(FIGURE_FILES)
	mkdir -p $(OUTPUT_DIRNAME)/ && \
	$(PANDOC) -f gfm -t latex --pdf-engine=xelatex -V geometry:margin=0.5in,bottom=0.8in -V block-headings -o $(PANDOC_DST)$@ $(patsubst %,$(PANDOC_SRC)%,$(DOC_FILES))
	ls -sh $(realpath $@)

$(OUTPUT_DIRNAME)/$(DOC_FILENAME).html: header.html $(DOC_FILES) $(FIGURE_FILES)
	mkdir -p $(OUTPUT_DIRNAME)/ && \
	cp -ap img/ $(shell pwd)/$(OUTPUT_DIRNAME)/&& \
	$(PANDOC) -f gfm -t html5 -H $(PANDOC_SRC)header.html --standalone -o $(PANDOC_DST)$@ $(patsubst %,$(PANDOC_SRC)%,$(DOC_FILES))
	ls -sh $(realpath $@)
endif

header.html: .tool/genheader.go specs-go/version.go
	rm -f go.mod go.sum && \
    go mod init github.com/opencontainers/distribution-spec && \
	go get github.com/opencontainers/distribution-spec/specs-go && \
	go run .tool/genheader.go > $@

install.tools: .install.gitvalidation

.install.gitvalidation:
	go install github.com/vbatts/git-validation@latest

conformance: conformance-test conformance-binary

conformance-test:
	$(GOLANGCILINT) -c 'cd conformance && golangci-lint run -v'

conformance-binary: $(OUTPUT_DIRNAME)/conformance.test

TEST_REGISTRY_CONTAINER ?= ghcr.io/project-zot/zot-minimal-linux-amd64:v2.0.0-rc6@sha256:bf95a94849cd9c6f596fb10e5a2d03b74267e7886d1ba0b3dab33337d9e46e5c
registry-ci:
	docker rm -f oci-conformance && \
		mkdir -p $(OUTPUT_DIRNAME) && \
		echo '{"distSpecVersion":"1.1.0-dev","storage":{"rootDirectory":"/tmp/zot","gc":false,"dedupe":false},"http":{"address":"0.0.0.0","port":"5000"},"log":{"level":"debug"}}' > $(shell pwd)/$(OUTPUT_DIRNAME)/zot-config.json
		docker run -d \
			-v $(shell pwd)/$(OUTPUT_DIRNAME)/zot-config.json:/etc/zot/config.json \
			--name=oci-conformance \
			-p 5000:5000 \
			$(TEST_REGISTRY_CONTAINER) && \
		sleep 5

conformance-ci:
	export OCI_ROOT_URL="http://localhost:5000" && \
		export OCI_NAMESPACE="myorg/myrepo" && \
		export OCI_TEST_PULL=1 && \
		export OCI_TEST_PUSH=1 && \
		export OCI_TEST_CONTENT_DISCOVERY=1 && \
		export OCI_TEST_CONTENT_MANAGEMENT=1 && \
		$(shell pwd)/$(OUTPUT_DIRNAME)/conformance.test

$(OUTPUT_DIRNAME)/conformance.test:
	cd conformance && \
		CGO_ENABLED=0 go test -c -o $(shell pwd)/$(OUTPUT_DIRNAME)/conformance.test \
			--ldflags="-X github.com/opencontainers/distribution-spec/conformance.Version=$(CONFORMANCE_VERSION)"
