# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux GODEBUG=madvdontneed=1,gctrace=0  go build -gcflags="all=-l -B" -ldflags="-s -w -buildid=" -trimpath -o main .
RUN go build -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 3000

ENV GOGC 1000
ENV GOMAXPROCS 3

CMD ["./main"]
