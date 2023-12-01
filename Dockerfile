FROM golang:1.20-alpine as builder
ARG TARGETARCH
COPY . .
RUN mkdir -p /app \
  && unset GOPATH \
  && GOOS=linux GOARCH=${TARGETARCH} go build -o /app/jetson-exporter

FROM waggle/plugin-base:1.1.1-base
COPY --from=builder /app/ /app/
WORKDIR /app
CMD ["/app/jetson-exporter"]