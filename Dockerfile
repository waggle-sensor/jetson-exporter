FROM golang:1.20 as builder
ARG TARGETARCH
COPY . .
RUN mkdir -p /app \
  && unset GOPATH \
  && CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o /app/jetson-exporter

FROM waggle/plugin-base:1.1.1-base

RUN apt-get update \
  && apt-get install -y \
  gnupg \
  ca-certificates \
  nano

COPY etc/apt/sources.list.d/nvidia-l4t-apt-source.list \
  /etc/apt/sources.list.d/nvidia-l4t-apt-source.list
RUN apt-key adv --fetch-key http://repo.download.nvidia.com/jetson/jetson-ota-public.asc \
  && mkdir -p /opt/nvidia/l4t-packages/ \
  && touch /opt/nvidia/l4t-packages/.nv-l4t-disable-boot-fw-update-in-preinstall \
  && apt-get update \
  && apt-get install --no-install-recommends -y \
  nvidia-l4t-tools

COPY --from=builder /app/ /app/
WORKDIR /app
CMD ["/app/jetson-exporter"]