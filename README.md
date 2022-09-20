# Jetson Exporter
Jetson exporter is a metric provider for Jetson Tegra GPU. Scrapers can hit `/metrics` endpoint to get Prometheus-formatted metrics. 

# Kubernetes
The exporter can be deployed as Kubernetes DaemonSet to provide the metrics per Jetson device.

# Main Advantage
Current Jetson platform for CUDA GPU (Sep 2022) is implemented differently from Desktop (amd64) CUDA platform. This blocks Jetson users from taking full features of Nvidia tools for device monitoring. `tegrastats` only provides a snapshot of GPU utilization which also makes users difficult to monitor usage while running CUDA-enabled programs. This exporter aggregates GPU utilization and provides wider picture of how CUDA GPU performs.

# Limitation
- Jetson GPU shares memory with CPU such that this exporter does not provide GPU memory usage
- We have not found a way to map GPU utilization with a process ID to identify which process is using the resource. This means that GPU utilization does not necessarily come from a particular program, but could come from other program running at the same time.

# Developer Note
Current provided metrics are limited to a few metrics. More metrics may be added if there are needs.