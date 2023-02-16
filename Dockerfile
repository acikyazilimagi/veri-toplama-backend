FROM golang:1.20-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -ldflags="-s -w" -tags=musl -o /app/api ./cmd/app


FROM gcr.io/distroless/static-debian11 as runner

COPY --from=builder --chown=nonroot:nonroot /app/api /

EXPOSE 80

ENTRYPOINT ["/api"]
