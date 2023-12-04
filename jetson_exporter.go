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
	flag.Parse()
	tegrastats := NewTegraStats()

	log.Println("Jetson exporter starts...")
	log.Println("Parameters are:")
	log.Printf("\t Endpoint: %s\n", metricsPath)
	log.Printf("\t TegraStats command: %v", tegrastats.GetTegraStatsCommandWithArguments())

	// watch signals to terminate external programs cleanly.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	log.Println("Executing the tegrastats command in the background...")
	err := tegrastats.Start()
	if err != nil {
		panic(err)
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(tegrastats)
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
			tegrastats.Close()
			return
		case <-sigc:
			log.Printf("OS signal received. Gracefully terminating...")
			tegrastats.Close()
			return
		}
	}
}
