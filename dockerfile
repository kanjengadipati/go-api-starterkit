FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app .

# final image
FROM alpine:latest

RUN apk add --no-cache ca-certificates wget \
	&& addgroup -S appgroup \
	&& adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/app .

EXPOSE 8080

USER appuser

HEALTHCHECK --interval=10s --timeout=5s --retries=5 CMD wget -qO- http://127.0.0.1:8080/health >/dev/null 2>&1 || exit 1

CMD ["./app"]
