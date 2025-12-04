FROM golang:1.24-alpine AS build

COPY . .
RUN CGO_ENABLED=0 go build -o /usr/local/bin/conformance .
ENTRYPOINT [ "/usr/local/bin/conformance" ]

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/local/bin/conformance /conformance
ENTRYPOINT [ "/conformance" ]
