FROM golang:1.21-alpine as builder
WORKDIR /app
COPY main.go .
RUN go build -o healthcheck main.go

FROM alpine:latest
COPY --from=builder /app/healthcheck /healthcheck
CMD ["/healthcheck"]