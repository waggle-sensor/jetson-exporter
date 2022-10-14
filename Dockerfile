FROM golang:1.17-alpine as builder
ARG TARGETARCH
COPY . .
RUN mkdir -p /app \
  && unset GOPATH \
  && GOOS=linux GOARCH=${TARGETARCH} go build -o /app/jetson-exporter

FROM golang:1.17-alpine
COPY --from=builder /app/ /app/
WORKDIR /app
CMD ["/app/jetson-exporter"]