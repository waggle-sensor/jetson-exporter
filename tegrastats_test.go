package main

import (
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestParseTegraStatsXavierNXJetPack411(t *testing.T) {
	testLine := "RAM 5246/7771MB (lfb 7x4MB) SWAP 983/20269MB (cached 280MB) CPU [47%@1420,23%@1420,32%@1420,22%@1420,31%@1420,96%@1420] EMC_FREQ 2%@1600 GR3D_FREQ 0%@1109 APE 150 MTS fg 1% bg 9% AO@29C GPU@31.5C PMIC@100C AUX@30C CPU@33.5C thermal@31.35C VDD_IN 6140/5510 VDD_CPU_GPU_CV 2706/2119 VDD_SOC 1074/1051"
	c := NewTegraStats()
	// Because the metrics include the time that represent when the stats collected,
	// we inject an arbitrary number that matches with the metrics file.
	c.AddTegraStatsStringRaw(testLine, time.Unix(int64(1701465532), 0))
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	metricsFile := "data/test_tegrastats_nx_jp411.txt"
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

func TestParseTegraStatsXavierNXJetPack512(t *testing.T) {
	testLine := "07-09-2024 13:31:31 RAM 3168/6851MB (lfb 2x4MB) SWAP 161/3426MB (cached 0MB) CPU [23%@1397,17%@1192,16%@1190,13%@1190,23%@1190,10%@1190] EMC_FREQ 2%@204 GR3D_FREQ 0%@[114] VIC_FREQ 115 APE 150 AUX@45C CPU@46.5C thermal@45.6C AO@45C GPU@45.5C PMIC@50C VDD_IN 2763mW/2846mW VDD_CPU_GPU_CV 650mW/670mW VDD_SOC 650mW/711mW"
	c := NewTegraStats()
	// Because the metrics include the time that represent when the stats collected,
	// we inject an arbitrary number that matches with the metrics file.
	c.AddTegraStatsStringRaw(testLine, time.Unix(int64(1701465532), 0))
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	metricsFile := "data/test_tegrastats_nx_jp512.txt"
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

func TestParseTegraStatsXavierNanoJetPack461(t *testing.T) {
	testLine := "RAM 1923/3964MB (lfb 2x2MB) SWAP 191/1982MB (cached 6MB) IRAM 0/252kB(lfb 252kB) CPU [5%@1224,4%@1224,3%@1224,6%@1224] EMC_FREQ 0%@1600 GR3D_FREQ 0%@76 APE 25 PLL@37C CPU@40C PMIC@50C GPU@40C AO@45C thermal@40C POM_5V_IN 1750/1628 POM_5V_GPU 0/0 POM_5V_CPU 366/406"
	c := NewTegraStats()
	// Because the metrics include the time that represent when the stats collected,
	// we inject an arbitrary number that matches with the metrics file.
	c.AddTegraStatsStringRaw(testLine, time.Unix(int64(1701465532), 0))
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	metricsFile := "data/test_tegrastats_xaviernano_jp461.txt"
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
