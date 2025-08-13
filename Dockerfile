FROM golang:1.25-alpine AS build_base
WORKDIR /tmp/awair-exporter

ARG VERSION="devel"

COPY . .

RUN go mod tidy && \
    go mod vendor && \
    CGO_ENABLED=0 go build \
        -ldflags="-s -w -X main.version=${VERSION}" \
        -o ./out/awair-exporter \
         cmd/awair-exporter/awair-exporter.go

FROM scratch
COPY --from=build_base /tmp/awair-exporter/out/awair-exporter /bin/awair-exporter
EXPOSE 8080/tcp
ENTRYPOINT ["/bin/awair-exporter"]