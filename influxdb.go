package main

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type PublisherConfig struct {
	NodeName                string
	InfluxDBURL             string
	InfluxDBToken           string
	InfluxDBPublishInterval int
	InfluxDBOrganization    string
	InfluxDBBucket          string
}

type Metrics struct {
	averagedLoad1s  float64
	averagedLoad5s  float64
	averagedLoad15s float64
	t               time.Time
}

type InfluxDBPublisher struct {
	config    PublisherConfig
	client    influxdb2.Client
	collector *TegraGPUCollector
}

func NewInfluxDBPublisher(pc PublisherConfig, c *TegraGPUCollector) *InfluxDBPublisher {
	return &InfluxDBPublisher{
		config:    pc,
		collector: c,
	}
}

func (p *InfluxDBPublisher) RunUntil(stopCh <-chan (bool)) {
	p.client = influxdb2.NewClient(p.config.InfluxDBURL, p.config.InfluxDBToken)
	ticker := time.NewTicker(time.Duration(p.config.InfluxDBPublishInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			m := Metrics{}
			p.collector.GetMetrics(&m)
			p.publishMetrics(m)
		case <-stopCh:
			return
		}
	}
}

func (p *InfluxDBPublisher) getInfluxDBClient() influxdb2.Client {
	if p.client.APIClient() == nil {
		p.client = influxdb2.NewClient(
			p.config.InfluxDBURL,
			p.config.InfluxDBToken)
	}
	return p.client
}

func (p *InfluxDBPublisher) publishMetrics(m Metrics) {
	c := p.getInfluxDBClient()
	api := c.WriteAPI(p.config.InfluxDBOrganization, p.config.InfluxDBBucket)
	p1 := p.createInfluxDBPoint("sys.metrics.gpu.average.1s", m.averagedLoad1s)
	p2 := p.createInfluxDBPoint("sys.metrics.gpu.average.5s", m.averagedLoad5s)
	p3 := p.createInfluxDBPoint("sys.metrics.gpu.average.15s", m.averagedLoad15s)
	api.WritePoint(p1)
	api.WritePoint(p2)
	api.WritePoint(p3)
	api.Flush()
}

func (p *InfluxDBPublisher) createInfluxDBPoint(k string, v interface{}) *write.Point {
	return influxdb2.NewPointWithMeasurement(k).
		AddTag("host", p.config.NodeName).
		AddField("_value", v).
		SetTime(time.Now().UTC())
}
