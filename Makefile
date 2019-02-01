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

# Generate json-schemas.
.PHONY: schemas
schemas:
	find schema -path 'schema/test-fixtures' -prune -o -name generate.jsonnet -execdir jsonnet -J jsonnet -m . {} \;

.PHONY: embed-schemas
embed-schemas:
	printf "%s\n\n%s\n" "$$(cat .header)" "$$(go generate schema/gen.go)" > schema/fs.go

.PHONY: test-fixtures
test-fixtures:
	find schema -path 'schema/test-fixtures/**' -name generate.jsonnet -execdir jsonnet -J jsonnet -m . {} \;

.PHONY: test-schemas
test-schemas:
	go test -cover ./schema
