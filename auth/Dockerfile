FROM golang:1.21.5-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# dependencies
COPY ["auth/go.mod", "auth/go.sum", "./"]
RUN go mod download

# build
COPY auth ./
RUN go build -o ./bin/app cmd/app/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY --from=builder /usr/local/src/configs /configs
COPY --from=builder /usr/local/src/migrations /migrations

CMD ["/app"]