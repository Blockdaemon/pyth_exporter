FROM golang:1.17-alpine as builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-w -s" -o /app/pyth_exporter .

FROM alpine
RUN apk add ca-certificates
COPY --from=builder /app/pyth_exporter /pyth_exporter
ENTRYPOINT ["/pyth_exporter"]
