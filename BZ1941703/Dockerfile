FROM golang:1.16 AS build

WORKDIR /go/src
COPY . .
RUN CGO_ENABLED=0 go build -o /usr/bin/server server/*.go
RUN CGO_ENABLED=0 go build -o /usr/bin/client client/main.go

FROM scratch
COPY --from=build /usr/bin/server /usr/bin/server
COPY --from=build /usr/bin/client /usr/bin/client

CMD ["/usr/bin/server"]
