build:
	CGO_ENABLED=0 go build -o ./out/jetson-exporter .

build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./out/jetson-exporter .

