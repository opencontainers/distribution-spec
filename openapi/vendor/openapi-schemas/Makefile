.PHONY: test clean schemas embed-files test-fixtures update-jsonnet-libs
test: clean test-fixtures schemas embed-files
	go test -cover ./...

clean:
	rm -rf ./bin
	find . -path '**/test-fixtures/**' -name '*.json' -exec rm -f {} \;
	find . -path 'jsonnetfile.*' -prune -o -name '*.schema.json' -exec rm -f {} \;

# Generate json-schemas.
schemas:
	find . -path **/test-fixtures -prune -o -name generate.jsonnet -execdir jsonnet -J jsonnet -m . {} \;

embed-files:
	find . -name gen.go -execdir go generate {} \;

test-fixtures:
	find . -path '**/test-fixtures/**' -name generate.jsonnet -execdir jsonnet -J jsonnet -m . {} \;

update-jsonnet-libs:
	jb update --jsonnetpkg-home=jsonnet
