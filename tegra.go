package main

import (
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// These paths are tested under Jetson Nano and NX on Sep 2022
type TegraGPUCollectorConfig struct {
	CollectionIntervalInMilli int
	LoadPath                  string
	CurrentDeviceFrqPathRex   string
}

// Interpretation comes from https://en.wikipedia.org/wiki/Load_(computing)
type TegraGPUCollector struct {
	mu              sync.Mutex
	config          *TegraGPUCollectorConfig
	deviceFrqPath   string
	currentLoad     float64
	averagedLoad1s  float64
	averagedLoad5s  float64
	averagedLoad15s float64
	load1sCoeff     float64
	load5sCoeff     float64
	load15sCoeff    float64

	descCurrentLoad *prometheus.Desc
	descLoad1s      *prometheus.Desc
	descLoad5s      *prometheus.Desc
	descLoad15s     *prometheus.Desc
}

func NewTegraGPUCollector(config *TegraGPUCollectorConfig) *TegraGPUCollector {
	return &TegraGPUCollector{
		config: config,
	}
}

func (c *TegraGPUCollector) Configure() error {
	c.load1sCoeff = 1. / math.Exp(float64(c.config.CollectionIntervalInMilli)/1000.)   // sampling frequency over a second
	c.load5sCoeff = 1. / math.Exp(float64(c.config.CollectionIntervalInMilli)/5000.)   // sampling frequency over 5 seconds
	c.load15sCoeff = 1. / math.Exp(float64(c.config.CollectionIntervalInMilli)/15000.) // sampling frequency over 15 seconds
	c.descCurrentLoad = prometheus.NewDesc(
		"gpu_current_load",
		"Current GPU load",
		nil, nil)
	c.descLoad1s = prometheus.NewDesc(
		"gpu_average_load1s",
		"Averaged GPU load in last second",
		nil, nil)
	c.descLoad5s = prometheus.NewDesc(
		"gpu_average_load5s",
		"Averaged GPU load in last 5 seconds",
		nil, nil)
	c.descLoad15s = prometheus.NewDesc(
		"gpu_average_load15s",
		"Averaged GPU load in last 15 seconds",
		nil, nil)
	return nil
}

func (c *TegraGPUCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.descCurrentLoad
	ch <- c.descLoad1s
	ch <- c.descLoad5s
	ch <- c.descLoad15s
}

func (c *TegraGPUCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.descCurrentLoad,
		prometheus.GaugeValue,
		roundFloat(c.currentLoad, 2),
	)
	ch <- prometheus.MustNewConstMetric(
		c.descLoad1s,
		prometheus.GaugeValue,
		roundFloat(c.averagedLoad1s, 2),
	)
	ch <- prometheus.MustNewConstMetric(
		c.descLoad5s,
		prometheus.GaugeValue,
		roundFloat(c.averagedLoad5s, 2),
	)
	ch <- prometheus.MustNewConstMetric(
		c.descLoad15s,
		prometheus.GaugeValue,
		roundFloat(c.averagedLoad5s, 2),
	)
}

func (c *TegraGPUCollector) GetMetrics(m *Metrics) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m.averagedLoad1s = c.averagedLoad1s
	m.averagedLoad5s = c.averagedLoad5s
	m.averagedLoad15s = c.averagedLoad15s
	m.t = time.Now().UTC()
}

func (c *TegraGPUCollector) RunUntil(stopCh <-chan (bool)) {
	ticker := time.NewTicker(time.Duration(c.config.CollectionIntervalInMilli) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			c.readMetrics()
		case <-stopCh:
			return
		}
	}
}

func (c *TegraGPUCollector) readMetrics() {
	if v, err := os.ReadFile(c.config.LoadPath); err == nil {
		value, err := strconv.ParseFloat(strings.Trim(string(v), "\n"), 64)
		if err == nil {
			c.mu.Lock()
			// load is ranged from 0 to 1000
			load := value / 1000.
			calcLoad(&c.averagedLoad1s, c.load1sCoeff, load)
			calcLoad(&c.averagedLoad5s, c.load5sCoeff, load)
			calcLoad(&c.averagedLoad15s, c.load15sCoeff, load)
			c.mu.Unlock()
		}
	}
}

func calcLoad(load *float64, exp float64, n float64) {
	*load *= exp
	*load += n * (1. - exp)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
