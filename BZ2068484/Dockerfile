FROM golang:alpine AS builder
WORKDIR /go/src
COPY server.go .
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/server /usr/bin/server
ENTRYPOINT ["/usr/bin/server"]
