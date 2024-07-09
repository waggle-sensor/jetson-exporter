package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Acknowledgement: the expressions and parser come from
// https://github.com/rbonghi/jetson_stats/blob/master/jtop/core/tegra_parse.py
// Detailed description on how to intepret the tegrastats output
// https://docs.nvidia.com/drive/drive_os_5.1.6.1L/nvvib_docs/index.html#page/DRIVE_OS_Linux_SDK_Development_Guide/Utilities/util_tegrastats.html
var (
	regSwap      = regexp.MustCompile(`SWAP (\d+)\/(\d+)(\w)B( ?)\(cached (\d+)(\w)B\)`)
	regCPU       = regexp.MustCompile(`CPU \[(.*?)\]`)
	regValueFreq = regexp.MustCompile(`\b(\d+)%@(\d+)`)
	regRAM       = regexp.MustCompile(`RAM (\d+)\/(\d+)(\w)B( ?)\(lfb (\d+)x(\d+)(\w)B\)`)
	regEMC       = regexp.MustCompile(`EMC_FREQ \b(\d+)%@(\d+)`)
	regMTS       = regexp.MustCompile(`MTS fg (\d+)% bg (\d+)%`)
	regGPU       = regexp.MustCompile(`GR3D_FREQ \b(\d+)%@[[]*(\d+)[]]*`)
	regWatt      = regexp.MustCompile(`\b(\w+) ([0-9.]+)[mW]*\/([0-9.]+)[mW]*\b`)
	regTemp      = regexp.MustCompile(`\b(\w+)@(-?[0-9.]+)C\b`)

	gBytes = 1024 * 1024 * 1024
	mBytes = 1024 * 1024
	kBytes = 1024
)

func parseUnit(s string) int {
	if s == "M" {
		return mBytes
	} else if s == "G" {
		return gBytes
	} else if s == "K" {
		return kBytes
	} else {
		return 1
	}
}

func parseValueWithFreq(s string) (v int, f int) {
	if m := regValueFreq.FindAllStringSubmatch(s, len(s)); m != nil {
		if value, err := strconv.Atoi(m[0][1]); err == nil {
			v = value
		}
		if freq, err := strconv.Atoi(m[0][2]); err == nil {
			f = freq
		}
	}
	return
}

type TegraStats struct {
	cmd                   *exec.Cmd
	TegraStatsString      string
	TegraStatsLastUpdated time.Time
	m                     map[string]*prometheus.Desc
	mu                    sync.Mutex
}

func NewTegraStats() *TegraStats {
	return NewTegraStatsWithCommand(exec.Command("tegrastats", "--interval", "2000"))
}

func NewTegraStatsWithCommand(c *exec.Cmd) *TegraStats {
	newTegraStats := &TegraStats{
		cmd: c,
		m: map[string]*prometheus.Desc{
			"tegra_last_updated_timestamp_epoch": prometheus.NewDesc(
				"tegra_last_updated_timestamp_epoch",
				"An epoch time of when the stats were collected from the system", nil, nil,
			),
			"tegra_temperature_celcius": prometheus.NewDesc(
				"tegra_temperature_celcius",
				"Temperature reading in Celcius", nil, nil,
			),
			"tegra_cpu_frequency_hz": prometheus.NewDesc(
				"tegra_cpu_frequency_hz",
				"CPU Clock frequency", nil, nil,
			),
			"tegra_cpu_util_percentage": prometheus.NewDesc(
				"tegra_cpu_util_percentage",
				"Utilization of CPU in percentage", nil, nil,
			),
			"tegra_emc_frequency_hz": prometheus.NewDesc(
				"tegra_emc_frequency_hz",
				"External memory controller clock frequency", nil, nil,
			),
			"tegra_emc_util_percentage": prometheus.NewDesc(
				"tegra_emc_util_percentage",
				"Utilization of external memory controller in percentage", nil, nil,
			),
			"tegra_gpu_frequency_hz": prometheus.NewDesc(
				"tegra_gpu_frequency_hz",
				"GPU clock frequency", nil, nil,
			),
			"tegra_gpu_util_percentage": prometheus.NewDesc(
				"tegra_gpu_util_percentage",
				"Utilization of GPU in percentage", nil, nil,
			),
			"tegra_lfb_nblock_count": prometheus.NewDesc(
				"tegra_lfb_nblock_count",
				"Count of largest free block", nil, nil,
			),
			"tegra_lfb_size_bytes": prometheus.NewDesc(
				"tegra_lfb_size_bytes",
				"Size of largest free block", nil, nil,
			),
			"tegra_mts_bg_percentage": prometheus.NewDesc(
				"tegra_mts_bg_percentage",
				"Time spent in foreground tasks", nil, nil,
			),
			"tegra_mts_fg_percentage": prometheus.NewDesc(
				"tegra_mts_fg_percentage",
				"Time spent in background tasks", nil, nil,
			),
			"tegra_ram_total_bytes": prometheus.NewDesc(
				"tegra_ram_total_bytes",
				"Total memory", nil, nil,
			),
			"tegra_ram_used_bytes": prometheus.NewDesc(
				"tegra_ram_used_bytes",
				"Current used memory", nil, nil,
			),
			"tegra_swap_total_bytes": prometheus.NewDesc(
				"tegra_swap_total_bytes",
				"Total swap memory", nil, nil,
			),
			"tegra_swap_cached_bytes": prometheus.NewDesc(
				"tegra_swap_cached_bytes",
				"Current swap cache memory", nil, nil,
			),
			"tegra_swap_used_bytes": prometheus.NewDesc(
				"tegra_swap_used_bytes",
				"Current swap used memory", nil, nil,
			),
			"tegra_wattage_current_milliwatts": prometheus.NewDesc(
				"tegra_wattage_current_milliwatts",
				"Current Watts of the hardware", nil, nil,
			),
			"tegra_wattage_average_milliwatts": prometheus.NewDesc(
				"tegra_wattage_average_milliwatts",
				"Averaged Watts of the hardware", nil, nil,
			),
		},
	}
	return newTegraStats
}

func (t *TegraStats) GetTegraStatsCommandWithArguments() []string {
	return t.cmd.Args
}

func (t *TegraStats) parseTegraStats(s string) []prometheus.Metric {
	metrics := []prometheus.Metric{}
	// RAM
	if m := regRAM.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[RAM 5246/7771MB (lfb 7x4MB) 5246 7771 M   7 4 M]]
		unitRAM := parseUnit(m[0][3])
		if ramUsed, err := strconv.Atoi(m[0][1]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_ram_used_bytes"],
				prometheus.GaugeValue,
				float64(ramUsed*unitRAM),
			))
		}
		if ramTotal, err := strconv.Atoi(m[0][2]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_ram_total_bytes"],
				prometheus.GaugeValue,
				float64(ramTotal*unitRAM),
			))
		}
		if lfbBlockCount, err := strconv.Atoi(m[0][5]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_lfb_nblock_count"],
				prometheus.GaugeValue,
				float64(lfbBlockCount),
			))
		}
		unitlfb := parseUnit(m[0][7])
		if blockSize, err := strconv.Atoi(m[0][6]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_lfb_size_bytes"],
				prometheus.GaugeValue,
				float64(blockSize*unitlfb),
			))
		}
	}

	// SWAP
	if m := regSwap.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[SWAP 983/20269MB (cached 280MB) 983 20269 M   280 M]]
		unitSwap := parseUnit(m[0][3])
		if swapUsed, err := strconv.Atoi(m[0][1]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_swap_used_bytes"],
				prometheus.GaugeValue,
				float64(swapUsed*unitSwap),
			))
		}
		if swapTotal, err := strconv.Atoi(m[0][2]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_swap_total_bytes"],
				prometheus.GaugeValue,
				float64(swapTotal*unitSwap),
			))
		}
		unitCached := parseUnit(m[0][6])
		if swapCached, err := strconv.Atoi(m[0][5]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_swap_cached_bytes"],
				prometheus.GaugeValue,
				float64(swapCached*unitCached),
			))
		}
	}

	// CPU
	if m := regCPU.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[CPU [47%@1420,23%@1420,32%@1420,22%@1420,31%@1420,96%@1420] 47%@1420,23%@1420,32%@1420,22%@1420,31%@1420,96%@1420]]
		labels := []string{"cpu"}
		dUtil := prometheus.NewDesc("tegra_cpu_util_percentage", "Utilization of CPU in percentage", labels, nil)
		dFreq := prometheus.NewDesc("tegra_cpu_frequency_hz", "CPU Clock frequency", labels, nil)
		for i, v := range strings.Split(m[0][1], ",") {
			if v == "off" {
				continue
			}
			p, f := parseValueWithFreq(v)
			metrics = append(metrics, prometheus.MustNewConstMetric(
				dUtil,
				prometheus.GaugeValue,
				float64(p),
				fmt.Sprintf("%d", i+1),
			))
			metrics = append(metrics, prometheus.MustNewConstMetric(
				dFreq,
				prometheus.GaugeValue,
				// tegrastats outputs the frequency in MHz
				float64(f*1e6),
				fmt.Sprintf("%d", i+1),
			))
		}
	}

	// EMC_FREQ
	if m := regEMC.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[EMC_FREQ 2%@1600 2 1600]]
		if emcUsed, err := strconv.Atoi(m[0][1]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_emc_util_percentage"],
				prometheus.GaugeValue,
				float64(emcUsed),
			))
		}
		if emcFreq, err := strconv.Atoi(m[0][2]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_emc_frequency_hz"],
				prometheus.GaugeValue,
				// tegrastats outputs the frequency in MHz
				float64(emcFreq*1e6),
			))
		}
	}

	// GPU
	if m := regGPU.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[GR3D_FREQ 0%@1109 0 1109]]
		if gpuUsed, err := strconv.Atoi(m[0][1]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_gpu_util_percentage"],
				prometheus.GaugeValue,
				float64(gpuUsed),
			))
		}
		if gpuFreq, err := strconv.Atoi(m[0][2]); err == nil {
			// tegrastats outputs the frequency in MHz
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_gpu_frequency_hz"],
				prometheus.GaugeValue,
				// tegrastats outputs the frequency in MHz
				float64(gpuFreq*1e6),
			))
		}
	}

	// MTS
	if m := regMTS.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[MTS fg 1% bg 9% 1 9]]
		if fg, err := strconv.Atoi(m[0][1]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_mts_fg_percentage"],
				prometheus.GaugeValue,
				float64(fg),
			))
		}
		if bg, err := strconv.Atoi(m[0][2]); err == nil {
			metrics = append(metrics, prometheus.MustNewConstMetric(
				t.m["tegra_mts_bg_percentage"],
				prometheus.GaugeValue,
				float64(bg),
			))
		}
	}

	// Temperature
	if m := regTemp.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[AO@29C AO 29] [GPU@31.5C GPU 31.5] [PMIC@100C PMIC 100] [AUX@30C AUX 30] [CPU@33.5C CPU 33.5] [thermal@31.35C thermal 31.35]]
		labels := []string{"sensor"}
		dTemp := prometheus.NewDesc("tegra_temperature_celcius", "Temperature reading in Celcius", labels, nil)
		for _, v := range m {
			if temp, err := strconv.ParseFloat(v[2], 32); err == nil {
				metrics = append(metrics, prometheus.MustNewConstMetric(
					dTemp,
					prometheus.GaugeValue,
					temp,
					strings.ToLower(v[1]),
				))
			}
		}
	}

	// Watts
	if m := regWatt.FindAllStringSubmatch(s, len(s)); m != nil {
		// [[VDD_IN 6140/5510 VDD_IN 6140 5510] [VDD_CPU_GPU_CV 2706/2119 VDD_CPU_GPU_CV 2706 2119] [VDD_SOC 1074/1051 VDD_SOC 1074 1051]]
		labels := []string{"sensor"}
		dCurWatt := prometheus.NewDesc("tegra_wattage_current_milliwatts", "Current Watts of the hardware", labels, nil)
		dAvgWatt := prometheus.NewDesc("tegra_wattage_average_milliwatts", "Averaged Watts of the hardware", labels, nil)
		for _, v := range m {
			if wattCurrent, err := strconv.Atoi(v[2]); err == nil {
				metrics = append(metrics, prometheus.MustNewConstMetric(
					dCurWatt,
					prometheus.GaugeValue,
					float64(wattCurrent),
					strings.ToLower(v[1]),
				))
			}
			if wattAveraged, err := strconv.Atoi(v[3]); err == nil {
				metrics = append(metrics, prometheus.MustNewConstMetric(
					dAvgWatt,
					prometheus.GaugeValue,
					float64(wattAveraged),
					strings.ToLower(v[1]),
				))
			}
		}
	}
	return metrics
}

func (t *TegraStats) AddTegraStatsString(s string) {
	t.mu.Lock()
	t.TegraStatsString = s
	t.TegraStatsLastUpdated = time.Now()
	t.mu.Unlock()
}

func (t *TegraStats) AddTegraStatsStringRaw(s string, _t time.Time) {
	t.mu.Lock()
	t.TegraStatsString = s
	t.TegraStatsLastUpdated = _t
	t.mu.Unlock()
}

func (t *TegraStats) Start() error {
	if t.cmd == nil {
		return fmt.Errorf("no tegrastats command is provided")
	}
	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	// stderr, err := t.cmd.StderrPipe()
	// if err != nil {
	// 	return err
	// }
	err = t.cmd.Start()
	if err != nil {
		return err
	}
	log.Printf("tegrastats starts with the process %d", t.cmd.Process.Pid)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				log.Println(err)
			} else {
				t.AddTegraStatsString(scanner.Text())
			}
		}
	}()
	return nil
}

func (t *TegraStats) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range t.m {
		ch <- d
	}
}

func (t *TegraStats) Collect(ch chan<- prometheus.Metric) {
	t.mu.Lock()
	statsStr := t.TegraStatsString
	lastUpdated := t.TegraStatsLastUpdated
	t.mu.Unlock()
	//
	if statsStr == "" {
		log.Printf("Prometheus Collect called, but tegrastats not yet collected.")
		return
	}
	if cm := t.parseTegraStats(statsStr); len(cm) > 0 {
		ch <- prometheus.MustNewConstMetric(
			t.m["tegra_last_updated_timestamp_epoch"],
			prometheus.GaugeValue,
			float64(lastUpdated.Unix()),
		)
		for _, c := range cm {
			ch <- c
		}
	}
}

func (t *TegraStats) Close() {
	if t.cmd != nil {
		if t.cmd.Process != nil {
			log.Printf("Killing tegrastats process %d...", t.cmd.Process.Pid)
			t.cmd.Process.Kill()
		}
	}
}
