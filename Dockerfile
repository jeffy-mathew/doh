FROM golang:1.22-alpine AS builder
WORKDIR /go/src/app
COPY . .
RUN go build -o /go/src/app/doh main.go

FROM alpine
RUN mkdir /app
WORKDIR /app

COPY --from=builder /go/src/app/doh /app/doh
RUN chmod +x /app/doh
EXPOSE 8080
ENTRYPOINT ["/app/doh"]