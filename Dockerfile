FROM golang:1.20-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download

WORKDIR /app/cmd/app

RUN go build -ldflags="-s -w" -tags=musl -o /build/app

FROM alpine AS production

COPY --from=builder /build/app /app

EXPOSE 80

ENTRYPOINT ["/app"]
