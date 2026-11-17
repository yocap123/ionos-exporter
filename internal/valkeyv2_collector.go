package internal

import (
	"github.com/prometheus/client_golang/prometheus"
)

type ValkeyV2Collector struct {
	clusterMetric  *prometheus.GaugeVec
	descFetchError *prometheus.Desc
}

func NewValkeyV2Collector() *ValkeyV2Collector {
	return &ValkeyV2Collector{
		clusterMetric: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ionos_valkey_v2_cluster",
			Help: "ValkeyV2 cluster exposed by its id, name and version",
		}, []string{"id", "name", "version"}),
		descFetchError: prometheus.NewDesc(
			"ionos_valkey_v2_fetch_error",
			"Error during Valkey V2 metrics fetch",
			nil,
			nil,
		),
	}
}

func (collector *ValkeyV2Collector) Describe(ch chan<- *prometheus.Desc) {
	collector.clusterMetric.Describe(ch)
	ch <- collector.descFetchError
}

func (collector *ValkeyV2Collector) Collect(ch chan<- prometheus.Metric) {
	valkeyV2Mutex.RLock()
	defer valkeyV2Mutex.RUnlock()

	collector.clusterMetric.Reset()
	for _, cluster := range valkeyV2Clusters {
		collector.clusterMetric.WithLabelValues(cluster.ID, cluster.Name, cluster.Version).Set(1)
	}
	collector.clusterMetric.Collect(ch)

	ch <- prometheus.MustNewConstMetric(
		collector.descFetchError,
		prometheus.GaugeValue, valkeyV2FetchError,
	)
}
