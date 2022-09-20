package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/waggle-sensor/jetson-exporter/collector"
)

// loadPath                = "/tmp/load"
// currentDeviceFrqPathRex = "/sys/devices/gpu.0/devfreq/*/cur_freq"
// loadPath                = "/sys/devices/gpu.0/load"
// currentDeviceFrqPathRex = "/sys/devices/gpu.0/devfreq/*/cur_freq"

func main() {
	config := &collector.TegraGPUCollectorConfig{}
	flag.IntVar(&config.CollectionIntervalInMilli, "sampling", 100, "Sampling interval in milliseconds")
	flag.StringVar(&config.LoadPath, "loadpath", "/sys/devices/gpu.0/load", "Path to GPU load")
	flag.StringVar(&config.CurrentDeviceFrqPathRex, "devfreqpathrex", "/sys/devices/gpu.0/devfreq/*/cur_freq", "Path described in Regression to current frequency of GPU device")
	collector := collector.NewTegraGPUCollector(config)
	collector.Configure()
	stopCh := make(chan bool, 1)
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collector)
	go collector.RunUntil(stopCh)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
