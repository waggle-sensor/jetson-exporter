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
    "github.com/waggle-sensor/jetson-exporter/pkg/tegracollect"
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
    flag.Parse()

    config := &tegracollect.TegraGPUCollectorConfig{
        CollectionIntervalInMilli: 1000,
        LoadPath:                  "/sys/devices/gpu.0/load",
        CurrentDeviceFrqPathRex:   "",
    }

    collector := tegracollect.NewTegraGPUCollector(config)
    if err := collector.Configure(); err != nil {
        log.Fatalf("Failed to configure collector: %v", err)
    }

    log.Println("Jetson exporter starts...")
    log.Println("Parameters are:")
    log.Printf("\t Endpoint: %s\n", metricsPath)

    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGQUIT)

    stopCh := make(chan bool)
    go collector.RunUntil(stopCh)

    reg := prometheus.NewRegistry()
    reg.MustRegister(collectors.NewGoCollector())
    reg.MustRegister(collector)
    http.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))

    sige := make(chan error, 1)
    go func() {
        err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), nil)
        sige <- err
    }()

    for {
        select {
        case err := <-sige:
            log.Println("HTTP listener returned with an error")
            log.Printf("%s\n", err)
            stopCh <- true
            return
        case <-sigc:
            log.Printf("OS signal received. Gracefully terminating...")
            stopCh <- true
            return
        }
    }
}
