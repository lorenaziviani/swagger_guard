FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go mod tidy
RUN go build -o swagger_guard main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/swagger_guard /app/swagger_guard
COPY --from=builder /app/api-spec.yaml /app/api-spec.yaml
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/.env.example /app/.env.example

ENTRYPOINT ["/app/swagger_guard"]
CMD ["parse", "--file", "api-spec.yaml"]