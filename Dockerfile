FROM golang:1.17-alpine as builder
ARG TARGETARCH
COPY . .
RUN mkdir -p /app \
  && make build \
  && cp -r ./out/ /app/

FROM golang:1.17-alpine
COPY --from=builder /app/ /app/
WORKDIR /app
CMD ["/app/jetson-exporter"]