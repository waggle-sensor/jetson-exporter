build:
	GOOS=linux GOARCH=arm64 go build -o ./out/jetson-exporter jetson_exporter.go
