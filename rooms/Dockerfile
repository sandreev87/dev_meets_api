FROM golang:1.21.5-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# dependencies
COPY ["rooms/go.mod", "rooms/go.sum", "./"]
RUN go mod download

# build
COPY rooms ./
RUN go build -o ./bin/rooms cmd/app/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/rooms /
COPY --from=builder /usr/local/src/configs /configs
COPY --from=builder /usr/local/src/web /web

CMD ["/rooms"]