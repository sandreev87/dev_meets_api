FROM golang:1.21.5-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# dependencies
COPY ["gateway/go.mod", "gateway/go.sum", "./"]
RUN go mod download

# build
COPY gateway ./
RUN go build -o ./bin/app cmd/app/main.go
RUN go build -o ./bin/migrator cmd/migrator/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY --from=builder /usr/local/src/bin/migrator /
COPY --from=builder /usr/local/src/configs /configs

CMD ["/migrator"]
CMD ["/app"]