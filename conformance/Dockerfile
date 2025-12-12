# Copyright contributors to the Open Containers Distribution Specification
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ---
# Stage 1: Install certs and build conformance binary
# ---
FROM docker.io/golang:1.19.7-alpine3.17 AS builder
ARG VERSION=unknown
ARG GO_PKG=github.com/opencontainers/distribution-spec/conformance
RUN apk --update add git make ca-certificates && mkdir -p /go/src/${GO_PKG}
WORKDIR /go/src/${GO_PKG}
ADD . .
RUN CGO_ENABLED=0 go test -c -o /conformance.test --ldflags="-X ${GO_PKG}.Version=${VERSION}"

# ---
# Stage 2: Final image with nothing but certs & binary
# ---
FROM scratch AS final
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /conformance.test /conformance.test
ENTRYPOINT ["/conformance.test"]
