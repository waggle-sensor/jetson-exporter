package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/waggle-sensor/jetson-exporter/collector"
)

func getenv(key string, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func main() {
	var (
		nodeName    string
		metricsPath string
	)
	flag.StringVar(&nodeName, "nodename", getenv("KUBENODE", ""), "Sampling interval in milliseconds")
	config := &collector.TegraGPUCollectorConfig{}
	flag.IntVar(&config.CollectionIntervalInMilli, "sampling", 100, "Sampling interval in milliseconds")
	flag.StringVar(&config.LoadPath, "loadpath", "/sys/devices/gpu.0/load", "Path to GPU load")
	flag.StringVar(&config.CurrentDeviceFrqPathRex, "devfreqpathrex", "/sys/devices/gpu.0/devfreq/*/cur_freq", "Path described in Regression to current frequency of GPU device")
	if nodeName != "" {
		metricsPath = path.Join("/", nodeName, "/metrics")
	} else {
		metricsPath = "/metrics"
	}

	fmt.Printf("Jetson exporter started\n")
	fmt.Println("Parameters are:")
	fmt.Printf("\t Sampling Interval: %d millisecond\n", config.CollectionIntervalInMilli)
	fmt.Printf("\t Loadpath: %s\n", config.LoadPath)
	fmt.Printf("\t Endpoint: %s\n", metricsPath)
	collector := collector.NewTegraGPUCollector(config)
	collector.Configure()
	stopCh := make(chan bool, 1)
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collector)
	go collector.RunUntil(stopCh)
	http.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	log.Fatal(http.ListenAndServe("0.0.0.0:9100", nil))
}
