build:
	go build -o ./out/jetson-exporter jetson_exporter.go

build-arm64:
	GOOS=linux GOARCH=arm64 go build -o ./out/jetson-exporter jetson_exporter.go

