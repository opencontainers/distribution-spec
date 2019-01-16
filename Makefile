EPOCH_TEST_COMMIT	:= 91d6d8466e68f1efff7977b63ad6f48e72245e05

test: .gitvalidation

# When this is running in travis, it will only check the travis commit range
.gitvalidation:
	@command -v git-validation >/dev/null 2>/dev/null || (echo "ERROR: git-validation not found. Consider 'make install.tools' target" && false)
ifdef TRAVIS_COMMIT_RANGE
	git-validation -q -run DCO,short-subject,dangling-whitespace
else
	git-validation -v -run DCO,short-subject,dangling-whitespace -range $(EPOCH_TEST_COMMIT)..HEAD
endif

install.tools: .install.gitvalidation

.install.gitvalidation:
	go get -u github.com/vbatts/git-validation

.PHONY: openapi
openapi: generate-openapi validate-openapi

# Generate OpenAPI Specification.
.PHONY: generate-openapi
generate-openapi:
	jsonnet -J openapi/vendor -m . openapi/openapi.jsonnet

# Validate generated OpenAPI specification file.
.PHONY: validate-openapi
validate-openapi:
	swagger-cli validate openapi.json
