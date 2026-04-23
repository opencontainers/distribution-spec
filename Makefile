EPOCH_TEST_COMMIT	:= 91d6d8466e68f1efff7977b63ad6f48e72245e05
CURRENT_COMMIT	:= $(shell git log --format="%H" -n 1)
CONFORMANCE_VERSION ?= $(CURRENT_COMMIT)

DOCKER	?= $(shell command -v docker 2>/dev/null)
PANDOC	?= $(shell command -v pandoc 2>/dev/null)
GOLANGCILINT	?= $(shell command -v golangcli-lint 2>/dev/null)

OUTPUT_DIRNAME	?= output/
DOC_FILENAME	?= oci-distribution-spec

# pinned to a 3.1 release due to a regression, see https://github.com/jgm/pandoc/issues/10952 before upgrading
PANDOC_CONTAINER ?= docker.io/pandoc/latex:3.1@sha256:1cf54d9214a9b52de2f58cf5895cc596a5960711a54d7938dc72f2b23473caf3
ifeq "$(strip $(PANDOC))" ''
	ifneq "$(strip $(DOCKER))" ''
		PANDOC = $(DOCKER) run \
			--rm \
			-v "$(CURDIR)/:/workdir/:ro" \
			-v "$(CURDIR)/$(OUTPUT_DIRNAME)/:/workdir/$(OUTPUT_DIRNAME)/" \
			-u "$(shell id -u):$(shell id -g)" \
			--workdir /workdir \
			$(PANDOC_CONTAINER)
	endif
endif

GOLANGCILINT_CONTAINER ?= docker.io/golangci/golangci-lint:v2.11.3@sha256:e838e8ab68aaefe83e2408691510867ade9329c0e0b895a3fb35eb93d1c2a4ba
ifeq "$(strip $(GOLANGCILINT))" ''
	ifneq "$(strip $(DOCKER))" ''
		GOLANGCILINT = $(DOCKER) run \
			--rm \
			-v $(shell pwd)/:/input:ro \
			-e GOCACHE=/tmp/.cache \
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
	go work init ./specs-go ./.tool && \
	go run .tool/genheader.go > $@ && \
	rm go.work

install.tools: .install.gitvalidation

.install.gitvalidation:
	go install github.com/vbatts/git-validation@latest

conformance: conformance-test conformance-cmd

conformance-test:
	$(GOLANGCILINT) -c 'cd conformance && golangci-lint run -v'

conformance-binary: $(OUTPUT_DIRNAME)/conformance.test

conformance-cmd: $(OUTPUT_DIRNAME)/conformance

registry-ci: registry-ci-olareg

TEST_REGISTRY_IMAGE_OLAREG ?= ghcr.io/olareg/olareg:edge
registry-ci-olareg:
	docker rm -f oci-conformance-olareg && \
		docker run --rm -d \
			--name=oci-conformance-olareg \
			-p 5000 \
			$(TEST_REGISTRY_IMAGE_OLAREG) serve --store-type mem --api-delete --api-blob-delete --api-sparse-image --api-sparse-index && \
		sleep 2

TEST_REGISTRY_IMAGE_ZOT ?= ghcr.io/project-zot/zot-minimal-linux-amd64:v2.1.7@sha256:2114797f00696011f38cc94c72f5773c84b1036562df5034d05ea19075179ad1
registry-ci-zot:
	docker rm -f oci-conformance-zot && \
		mkdir -p $(OUTPUT_DIRNAME) && \
		echo '{"distSpecVersion":"1.1.0-dev","storage":{"rootDirectory":"/tmp/zot","gc":false,"dedupe":false},"http":{"address":"0.0.0.0","port":"5000"},"log":{"level":"debug"}}' > $(shell pwd)/$(OUTPUT_DIRNAME)/zot-config.json
		docker run --rm -d \
			-v $(shell pwd)/$(OUTPUT_DIRNAME)/zot-config.json:/etc/zot/config.json \
			--name=oci-conformance-zot \
			-p 5000 \
			$(TEST_REGISTRY_IMAGE_ZOT) && \
		sleep 5

conformance-ci: conformance-ci-olareg

conformance-ci-olareg: $(OUTPUT_DIRNAME)/conformance
	export OCI_VERSION="dev" && \
		export OCI_REGISTRY="localhost:$$(docker port oci-conformance-olareg 5000| head -1 | cut -f2 -d:)" && \
		export OCI_TLS="disabled" && \
		export OCI_REPO1="myorg/myrepo" && \
		export OCI_REPO2="myorg/myrepo2" && \
		export OCI_RESULTS_DIR="." && \
		export OCI_DATA_SPARSE=true && \
		$(shell pwd)/$(OUTPUT_DIRNAME)/conformance

conformance-ci-zot: $(OUTPUT_DIRNAME)/conformance
	export OCI_REGISTRY="localhost:$$(docker port oci-conformance-zot 5000| head -1 | cut -f2 -d:)" && \
		export OCI_TLS="disabled" && \
		export OCI_REPO1="myorg/myrepo" && \
		export OCI_REPO2="myorg/myrepo2" && \
		export OCI_RESULTS_DIR="." && \
		$(shell pwd)/$(OUTPUT_DIRNAME)/conformance

clean-ci:
	docker rm -f oci-conformance-olareg oci-conformance-zot

$(OUTPUT_DIRNAME)/conformance: conformance/*.go conformance/go.mod
	cd conformance && \
		CGO_ENABLED=0 go build -o $(shell pwd)/$(OUTPUT_DIRNAME)/conformance \
			--ldflags="-X github.com/opencontainers/distribution-spec/conformance.Version=$(CONFORMANCE_VERSION)"

$(OUTPUT_DIRNAME)/conformance.test: conformance/*.go conformance/go.mod
	cd conformance && \
		CGO_ENABLED=0 go test -c -o $(shell pwd)/$(OUTPUT_DIRNAME)/conformance.test \
			--ldflags="-X github.com/opencontainers/distribution-spec/conformance.Version=$(CONFORMANCE_VERSION)"

clean: clean-ci
	rm -rf header.html junit.xml report.html results.yaml output conformance/results
