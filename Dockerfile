FROM golang:1.22.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN go build -ldflags "-w -s" -o main

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /app/main ./

ENTRYPOINT ["/app/main"]