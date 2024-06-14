package internal

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type openStackCollector struct {
	collectDuration *prometheus.Desc
	// Compute metrics
	maxTotalCores          *prometheus.Desc
	maxTotalInstances      *prometheus.Desc
	maxTotalRAMSize        *prometheus.Desc
	perFlavorInstanceCount *prometheus.Desc
	perStatusInstanceCount *prometheus.Desc
	totalCoresUsed         *prometheus.Desc
	totalInstancesUsed     *prometheus.Desc
	totalRAMUsed           *prometheus.Desc
	// Volume metrics
	containerBytesUsed      *prometheus.Desc
	maxTotalVolumeGigabytes *prometheus.Desc
	maxTotalVolumes         *prometheus.Desc
	perStatusVolumeCount    *prometheus.Desc
	totalGigabytesUsed      *prometheus.Desc
	totalVolumesUsed        *prometheus.Desc
	volumeLimit             float64
}

func NewOpenStackCollector(volumeLimit float64) *openStackCollector {
	return &openStackCollector{
		collectDuration: prometheus.NewDesc("openstack_collect_duration_seconds",
			"The time it took to collect the metrics in seconds",
			nil, nil,
		),
		// Compute metrics
		maxTotalCores: prometheus.NewDesc("openstack_max_total_cores",
			"The limit of cores that can be assigned to instances in the project",
			nil, nil,
		),
		maxTotalInstances: prometheus.NewDesc("openstack_max_total_instances",
			"The limit of total instances in the project",
			nil, nil,
		),
		maxTotalRAMSize: prometheus.NewDesc("openstack_max_total_ram_size",
			"The limit of RAM that can be assigned to instances in the project",
			nil, nil,
		),
		perFlavorInstanceCount: prometheus.NewDesc("openstack_per_flavor_instance_count",
			"Number of instances per flavor",
			[]string{"flavor"}, nil,
		),
		perStatusInstanceCount: prometheus.NewDesc("openstack_per_status_instance_count",
			"Number of instances per status",
			[]string{"status"}, nil,
		),
		totalCoresUsed: prometheus.NewDesc("openstack_total_cores_used",
			"The current number of cores used",
			nil, nil,
		),
		totalInstancesUsed: prometheus.NewDesc("openstack_total_instances_used",
			"The current number of instances",
			nil, nil,
		),
		totalRAMUsed: prometheus.NewDesc("openstack_total_ram_used",
			"The current number RAM used",
			nil, nil,
		),
		// Volume metrics
		containerBytesUsed: prometheus.NewDesc("openstack_container_bytes_used",
			"The total of bytes stored in the container",
			[]string{"container"}, nil,
		),
		maxTotalVolumeGigabytes: prometheus.NewDesc("openstack_max_total_volume_gigabytes",
			"The limit of total volume size in the project",
			nil, nil,
		),
		maxTotalVolumes: prometheus.NewDesc("openstack_max_total_volumes",
			"The limit of total volumes in the project",
			nil, nil,
		),
		perStatusVolumeCount: prometheus.NewDesc("openstack_per_status_volume_count",
			"Number of volumes per status",
			[]string{"status"}, nil,
		),
		totalGigabytesUsed: prometheus.NewDesc("openstack_total_volume_gigabytes_used",
			"The current total of gigabytes used in volumes",
			nil, nil,
		),
		totalVolumesUsed: prometheus.NewDesc("openstack_total_volumes_used",
			"The current number of volumes",
			nil, nil,
		),
		volumeLimit: volumeLimit,
	}
}

func (c *openStackCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.collectDuration
	// Compute metrics
	ch <- c.maxTotalCores
	ch <- c.maxTotalInstances
	ch <- c.maxTotalRAMSize
	ch <- c.perFlavorInstanceCount
	ch <- c.perStatusInstanceCount
	ch <- c.totalCoresUsed
	ch <- c.totalInstancesUsed
	ch <- c.totalRAMUsed
	// Volume metrics
	ch <- c.containerBytesUsed
	ch <- c.maxTotalVolumeGigabytes
	ch <- c.maxTotalVolumes
	ch <- c.totalGigabytesUsed
	ch <- c.totalVolumesUsed
}

func (collector *openStackCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(logger).Log("msg", "Starting metrics collection", "address", ":9595")
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime).Seconds()
		level.Debug(logger).Log("msg", fmt.Sprintf("Metrics collection duration: %f seconds", duration))
		ch <- prometheus.MustNewConstMetric(collector.collectDuration, prometheus.GaugeValue, duration)
	}()
	providerClient, err := authenticateOpenStack()

	if err != nil {
		level.Error(logger).Log("message", "Failed to authenticate to OpenStack API", "err", err)
		return
	}

	computeLimits := getComputeLimits(providerClient)
	maxTotalCores := float64(computeLimits.Absolute.MaxTotalCores)
	maxTotalInstances := float64(computeLimits.Absolute.MaxTotalInstances)
	maxTotalRAMSize := float64(computeLimits.Absolute.MaxTotalRAMSize)
	totalCoresUsed := float64(computeLimits.Absolute.TotalCoresUsed)
	totalInstancesUsed := float64(computeLimits.Absolute.TotalInstancesUsed)
	totalRAMUsed := float64(computeLimits.Absolute.TotalRAMUsed)

	maxTotalCoresMetric := prometheus.MustNewConstMetric(collector.maxTotalCores, prometheus.GaugeValue, maxTotalCores)
	maxTotalInstancesMetric := prometheus.MustNewConstMetric(collector.maxTotalInstances, prometheus.GaugeValue, maxTotalInstances)
	maxTotalRAMSizeMetric := prometheus.MustNewConstMetric(collector.maxTotalRAMSize, prometheus.GaugeValue, maxTotalRAMSize)
	totalCoresUsedMetric := prometheus.MustNewConstMetric(collector.totalCoresUsed, prometheus.GaugeValue, totalCoresUsed)
	totalInstancesUsedMetric := prometheus.MustNewConstMetric(collector.totalInstancesUsed, prometheus.GaugeValue, totalInstancesUsed)
	totalRAMUsedMetric := prometheus.MustNewConstMetric(collector.totalRAMUsed, prometheus.GaugeValue, totalRAMUsed)

	serverList := getAllServers(providerClient)
	flavorCount := countInstancePerFlavor(serverList)
	for flavor, count := range flavorCount {
		flavorCountMetric := prometheus.MustNewConstMetric(collector.perFlavorInstanceCount, prometheus.GaugeValue, float64(count), flavor)
		ch <- flavorCountMetric
	}

	statusCountServers := countInstancePerStatus(serverList)
	for status, count := range statusCountServers {
		statusCountMetric := prometheus.MustNewConstMetric(collector.perStatusInstanceCount, prometheus.GaugeValue, float64(count), status)
		ch <- statusCountMetric
	}

	volumeList := getAllVolumes(providerClient)
	statusCountVolumes := countVolumePerStatus(volumeList)
	for status, count := range statusCountVolumes {
		statusCountMetric := prometheus.MustNewConstMetric(collector.perStatusVolumeCount, prometheus.GaugeValue, float64(count), status)
		ch <- statusCountMetric
	}

	if !strings.Contains(os.Getenv("OS_AUTH_URL"), "otc") {

		volumeLimits := getVolumeLimits(providerClient)
		maxTotalVolumeGigabytes := float64(volumeLimits.Absolute.MaxTotalVolumeGigabytes)
		maxTotalVolumes := float64(volumeLimits.Absolute.MaxTotalVolumes)
		totalGigabytesUsed := float64(volumeLimits.Absolute.TotalGigabytesUsed)
		totalVolumesUsed := float64(volumeLimits.Absolute.TotalVolumesUsed)

		maxTotalVolumeGigabytesMetric := prometheus.MustNewConstMetric(collector.maxTotalVolumeGigabytes, prometheus.GaugeValue, maxTotalVolumeGigabytes)
		maxTotalVolumesMetric := prometheus.MustNewConstMetric(collector.maxTotalVolumes, prometheus.GaugeValue, maxTotalVolumes)
		totalGigabytesUsedMetric := prometheus.MustNewConstMetric(collector.totalGigabytesUsed, prometheus.GaugeValue, totalGigabytesUsed)
		totalVolumesUsedMetric := prometheus.MustNewConstMetric(collector.totalVolumesUsed, prometheus.GaugeValue, totalVolumesUsed)
		// Volume metrics
		ch <- maxTotalVolumeGigabytesMetric
		ch <- maxTotalVolumesMetric
		ch <- totalGigabytesUsedMetric
		ch <- totalVolumesUsedMetric
	} else {
		totalVolumesUsed := float64(len(volumeList))
		maxTotalVolumes := float64(collector.volumeLimit)
		maxTotalVolumesMetric := prometheus.MustNewConstMetric(collector.maxTotalVolumes, prometheus.GaugeValue, maxTotalVolumes)
		totalVolumesUsedMetric := prometheus.MustNewConstMetric(collector.totalVolumesUsed, prometheus.GaugeValue, totalVolumesUsed)
		ch <- maxTotalVolumesMetric
		ch <- totalVolumesUsedMetric
	}

	containers := getContainerList(providerClient)
	for _, container := range containers {
		containerBytesUsedMetric := prometheus.MustNewConstMetric(collector.containerBytesUsed, prometheus.GaugeValue, float64(container.Bytes), container.Name)
		ch <- containerBytesUsedMetric
	}

	// Compute metrics
	ch <- maxTotalCoresMetric
	ch <- maxTotalInstancesMetric
	ch <- maxTotalRAMSizeMetric
	ch <- totalCoresUsedMetric
	ch <- totalInstancesUsedMetric
	ch <- totalRAMUsedMetric

	level.Info(logger).Log("msg", "Finished metrics collection", "address", ":9595")
}
