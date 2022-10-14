package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getenv(key string, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func main() {
	var port string
	metricsPath := "/metrics"
	flag.StringVar(&port, "port", getenv("PORT", "9091"), "Port number to listen")
	var collectorConfig TegraGPUCollectorConfig
	flag.IntVar(&collectorConfig.CollectionIntervalInMilli, "sampling", 100, "Sampling interval in milliseconds")
	flag.StringVar(&collectorConfig.LoadPath, "loadpath", "/sys/devices/gpu.0/load", "Path to GPU load")
	flag.StringVar(&collectorConfig.CurrentDeviceFrqPathRex, "devfreqpathrex", "/sys/devices/gpu.0/devfreq/*/cur_freq", "Path described in Regression to current frequency of GPU device")
	var publisherConfig PublisherConfig
	flag.StringVar(&publisherConfig.NodeName, "nodename", getenv("KUBENODE", ""), "Name of the Kubernetes node")
	flag.StringVar(&publisherConfig.InfluxDBURL, "influxdb-url", getenv("INFLUXDB_URL", ""), "InfluxDB URL")
	flag.StringVar(&publisherConfig.InfluxDBToken, "influxdb-token", getenv("INFLUXDB_TOKEN", ""), "InfluxDB token")
	flag.StringVar(&publisherConfig.InfluxDBOrganization, "influxdb-org", getenv("INFLUXDB_ORG", "waggle"), "InfluxDB organization")
	flag.StringVar(&publisherConfig.InfluxDBBucket, "influxdb-bucket", getenv("INFLUXDB_BUCKET", "waggle"), "InfluxDB bucket")
	flag.IntVar(&publisherConfig.InfluxDBPublishInterval, "influxdb-interval", 1, "InlufxDB publishing interval in seconds")
	fmt.Println("Jetson exporter started")
	fmt.Println("Parameters are:")
	fmt.Printf("\t Sampling Interval: %d millisecond\n", collectorConfig.CollectionIntervalInMilli)
	fmt.Printf("\t Loadpath: %s\n", collectorConfig.LoadPath)
	fmt.Printf("\t Endpoint: %s\n", metricsPath)
	collector := NewTegraGPUCollector(&collectorConfig)
	collector.Configure()
	stopCh := make(chan bool, 1)
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collector)
	go collector.RunUntil(stopCh)
	if publisherConfig.InfluxDBURL != "" {
		fmt.Println("InfluxDB URL is provided. Metrics will be published.")
		fmt.Printf("\t Publishing Interval: %d second(s) \n", publisherConfig.InfluxDBPublishInterval)
		publisher := NewInfluxDBPublisher(publisherConfig, collector)
		go publisher.RunUntil(stopCh)
	}
	http.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil))
}
