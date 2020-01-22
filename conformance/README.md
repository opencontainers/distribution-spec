## Conformance Tests

### How to Run

Requires Go 1.13+.

In this directory, build the test binary:
```
go test -c
```

This will produce an executable at `conformance.test`.

Next, set environment variables with your registry details:
```
export OCI_ROOT_URL="https://r.myreg.io"
export OCI_NAMESPACE="myorg/myrepo"
export OCI_USERNAME="myuser"
export OCI_PASSWORD="mypass"
export OCI_DEBUG="true"
```

Lastly, run the tests:
```
./conformance.test
```

This will produce `junit.xml` and `report.html` with the results.

Note: for some registries, you may need to create `OCI_NAMESPACE` ahead of time.
