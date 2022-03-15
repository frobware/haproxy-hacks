FROM golang:1.16 AS build

WORKDIR /go/src
COPY server.go .
RUN CGO_ENABLED=0 go build -o /usr/bin/server server.go

FROM scratch
COPY --from=build /usr/bin/server /usr/bin/server

CMD ["/usr/bin/server"]
