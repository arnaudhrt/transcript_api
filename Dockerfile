FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o transcript-api .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/transcript-api .

EXPOSE 8000
CMD ["./transcript-api"]
