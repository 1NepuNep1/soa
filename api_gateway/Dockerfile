FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy && go build -o gateway .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/gateway .
EXPOSE 8080
CMD ["./gateway"]

