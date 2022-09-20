package collector

import (
	"os"
	"testing"
	"time"
)

func TestTegraGPUCollector(t *testing.T) {
	config := &TegraGPUCollectorConfig{
		LoadPath: "/tmp/load",
	}
	err := os.WriteFile(config.LoadPath, []byte("0"), 0644)
	if err != nil {
		panic(err)
	}
	c := NewTegraGPUCollector(config)
	c.Configure()
	stopCh := make(chan bool, 1)
	go c.RunUntil(stopCh)
	for {
		t.Logf("Load %2.2f %2.2f %2.2f",
			c.averagedLoad1s,
			c.averagedLoad5s,
			c.averagedLoad15s,
		)
		time.Sleep(1 * time.Second)
	}
}
