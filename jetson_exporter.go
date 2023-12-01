package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	flag.StringVar(&publisherConfig.InfluxDBURL, "influxdb-url", getenv("INFLUXDB_URL", ""), "InfluxDB URL")
	flag.StringVar(&publisherConfig.InfluxDBToken, "influxdb-token", getenv("INFLUXDB_TOKEN", ""), "InfluxDB token")
	flag.StringVar(&publisherConfig.InfluxDBOrganization, "influxdb-org", getenv("INFLUXDB_ORG", "waggle"), "InfluxDB organization")
	flag.StringVar(&publisherConfig.InfluxDBBucket, "influxdb-bucket", getenv("INFLUXDB_BUCKET", "waggle"), "InfluxDB bucket")
	flag.IntVar(&publisherConfig.InfluxDBPublishInterval, "influxdb-interval", 1, "InlufxDB publishing interval in seconds")
	flag.Parse()
	fmt.Println("Jetson exporter started")
	fmt.Println("Parameters are:")
	fmt.Printf("\t Sampling Interval: %d millisecond\n", collectorConfig.CollectionIntervalInMilli)
	fmt.Printf("\t Loadpath: %s\n", collectorConfig.LoadPath)
	fmt.Printf("\t Endpoint: %s\n", metricsPath)
	collector := NewTegraStats()

	// watch signals to terminate external programs cleanly.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	err := collector.Start()
	if err != nil {
		panic(err)
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collector)
	// go collector.RunUntil(stopCh)
	http.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
	for {
		select {
		// case line := <-m:
		// 	// log.Println(line)
		// 	parseTegraStats(line)
		case <-sigc:
			log.Printf("terminating")
			collector.Close()
			return
		}
	}
}
