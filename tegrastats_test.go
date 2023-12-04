package main

import (
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestParseTegraStats(t *testing.T) {
	testLine := "RAM 5246/7771MB (lfb 7x4MB) SWAP 983/20269MB (cached 280MB) CPU [47%@1420,23%@1420,32%@1420,22%@1420,31%@1420,96%@1420] EMC_FREQ 2%@1600 GR3D_FREQ 0%@1109 APE 150 MTS fg 1% bg 9% AO@29C GPU@31.5C PMIC@100C AUX@30C CPU@33.5C thermal@31.35C VDD_IN 6140/5510 VDD_CPU_GPU_CV 2706/2119 VDD_SOC 1074/1051"
	c := NewTegraStats()
	// Because the metrics include the time that represent when the stats collected,
	// we inject an arbitrary number that matches with the metrics file.
	c.AddTegraStatsStringRaw(testLine, time.Unix(int64(1701465532), 0))
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	metricsFile := "data/test_tegrastats.txt"
	// NOTE: Uncomment below to generate a new file.
	//       Then, you will need to comment this out again for this unittest
	// err := prometheus.WriteToTextfile(metricsFile, reg)
	// if err != nil {
	// 	t.Fatalf("Metric comparison failed: %s", err)
	// }

	wantMetrics, err := os.Open(metricsFile)
	if err != nil {
		t.Fatalf("unable to read input test file %s", metricsFile)
	}
	err = testutil.GatherAndCompare(reg, wantMetrics)
	if err != nil {
		t.Fatalf("Metric comparison failed: %s", err)
	}
}
