FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/taskflow-api ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN adduser -D -H -u 10001 appuser \
	&& apk add --no-cache ca-certificates

COPY --from=builder /out/taskflow-api /app/taskflow-api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/taskflow-api"]
