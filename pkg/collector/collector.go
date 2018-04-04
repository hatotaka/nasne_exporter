package collector

import (
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne-exporter/pkg/nasneclient"
	"github.com/prometheus/client_golang/prometheus"
)

func NewNasneCollector(nasneAddrs []string) *NasneCollector {
	return &NasneCollector{
		nasneAddrs: nasneAddrs,

		info: prometheus.NewDesc(
			"nasne_info",
			"information of nasne",
			[]string{
				"name",
				"software_version",
				"hardware_version",
				"product_name",
			}, nil,
		),
		hddTotal: prometheus.NewDesc(
			"nasne_hdd_byte_total",
			"hdd byte of nasne",
			[]string{
				"name",
				"id",
				"format",
				"hdd_name",
				"vendor_id",
				"product_id",
			},
			nil,
		),
		hddUsed: prometheus.NewDesc(
			"nasne_hdd_byte_used",
			"hdd byte of nasne",
			[]string{
				"name",
				"id",
				"format",
				"hdd_name",
				"vendor_id",
				"product_id",
			},
			nil,
		),
		dtcpipClientTotal: prometheus.NewDesc(
			"nasne_dtcpip_client_total",
			"number of dtcpip client",
			[]string{
				"name",
			},
			nil,
		),
		recordedTitleTotal: prometheus.NewDesc(
			"nasne_recorded_title_total",
			"number of dtcpip client",
			[]string{
				"name",
			},
			nil,
		),
		reservedConflictTotal: prometheus.NewDesc(
			"nasne_conflict_total",
			"number of conflict",
			[]string{
				"name",
			},
			nil,
		),
		collectTime: prometheus.NewDesc(
			"nasne_last_collect_time",
			"time of last collect",
			nil, nil,
		),

		totalCollectionDurationsHistogram: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "nasne_total_collection_durations_histogram_seconds",
				Help:    "Total collection latency distributions.",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
			},
		),

		collectionDurationsHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "nasne_collection_durations_histogram_seconds",
				Help:    "Collection latency distributions.",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
			},
			[]string{"name"},
		),
	}
}

type NasneCollector struct {
	nasneAddrs []string

	info                  *prometheus.Desc
	hddTotal              *prometheus.Desc
	hddUsed               *prometheus.Desc
	dtcpipClientTotal     *prometheus.Desc
	recordedTitleTotal    *prometheus.Desc
	reservedConflictTotal *prometheus.Desc
	collectTime           *prometheus.Desc

	totalCollectionDurationsHistogram prometheus.Histogram
	collectionDurationsHistogram      *prometheus.HistogramVec

	cache metricsCache
}

type metricsCache struct {
	mu sync.Mutex
	m  []prometheus.Metric
}

func (mc *metricsCache) Set(metricsList []prometheus.Metric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.m = metricsList
}

func (mc *metricsCache) Get() []prometheus.Metric {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	return mc.m
}

func (n *NasneCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- n.info
	ch <- n.hddTotal
	ch <- n.hddUsed
	ch <- n.dtcpipClientTotal
	ch <- n.recordedTitleTotal
	ch <- n.reservedConflictTotal
	ch <- n.collectTime
}

func (n *NasneCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range n.cache.Get() {
		ch <- m
	}
}

func (n *NasneCollector) Run() error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	n.runCollect()

	for t := range ticker.C {
		glog.V(2).Info(t)

		n.runCollect()
	}

	return nil
}

func (n *NasneCollector) RegisterCollector(r *prometheus.Registry) {
	r.MustRegister(n)
	r.MustRegister(n.totalCollectionDurationsHistogram)
	r.MustRegister(n.collectionDurationsHistogram)
}

func (n *NasneCollector) collectNasneCollector(start, end time.Time) ([]prometheus.Metric, error) {
	n.totalCollectionDurationsHistogram.Observe(float64(end.Sub(start).Nanoseconds()) / float64(time.Second))

	return []prometheus.Metric{
		prometheus.MustNewConstMetric(n.collectTime, prometheus.GaugeValue, float64(start.Unix())),
	}, nil
}

func (n *NasneCollector) collectCollectionDuration(start, end time.Time, commonLabel []string) ([]prometheus.Metric, error) {
	n.collectionDurationsHistogram.WithLabelValues(commonLabel...).Observe(float64(end.Sub(start).Nanoseconds()) / float64(time.Second))
	return []prometheus.Metric{}, nil
}

func (n *NasneCollector) collectVersion(client *nasneclient.NasneClient, commonLabel []string) ([]prometheus.Metric, error) {
	softwareVersion, err := client.GetSoftwareVersion()
	if err != nil {
		return nil, err
	}

	hardwareVersion, err := client.GetHardwareVersion()
	if err != nil {
		return nil, err
	}

	labelValues := []string{}
	labelValues = append(labelValues, commonLabel...)
	labelValues = append(labelValues,
		softwareVersion.SoftwareVersion,
		strconv.Itoa(hardwareVersion.HardwareVersion),
		hardwareVersion.ProductName,
	)

	return []prometheus.Metric{
		prometheus.MustNewConstMetric(n.info, prometheus.GaugeValue, float64(1), labelValues...),
	}, nil
}

func (n *NasneCollector) collectHDD(client *nasneclient.NasneClient, commonLabel []string) ([]prometheus.Metric, error) {
	hddList, err := client.GetHDDList()
	if err != nil {
		return nil, err
	}

	metrics := []prometheus.Metric{}
	for _, hdd := range hddList.HDD {
		hddInfo, err := client.GetHDDInfo(hdd.ID)
		if err != nil {
			glog.Fatal(err)
		}

		labelValues := []string{}
		labelValues = append(labelValues, commonLabel...)
		labelValues = append(labelValues,
			strconv.Itoa(hddInfo.HDD.ID),
			hddInfo.HDD.Format,
			hddInfo.HDD.Name,
			hddInfo.HDD.VendorID,
			hddInfo.HDD.ProductID,
		)

		metrics = append(metrics,
			prometheus.MustNewConstMetric(n.hddTotal, prometheus.GaugeValue, hddInfo.HDD.TotalVolumeSize, labelValues...),
			prometheus.MustNewConstMetric(n.hddUsed, prometheus.GaugeValue, hddInfo.HDD.UsedVolumeSize, labelValues...),
		)
	}

	return metrics, nil
}

func (n *NasneCollector) collectDTCPClient(client *nasneclient.NasneClient, commonLabel []string) ([]prometheus.Metric, error) {
	dtcpipClientList, err := client.GetDTCPIPClientList()
	if err != nil {
		return nil, err
	}

	return []prometheus.Metric{
		prometheus.MustNewConstMetric(n.dtcpipClientTotal, prometheus.GaugeValue, float64(dtcpipClientList.Number), commonLabel...),
	}, nil
}

func (n *NasneCollector) collectRecord(client *nasneclient.NasneClient, commonLabel []string) ([]prometheus.Metric, error) {
	recordedTitleList, err := client.GetRecordedTitleList()
	if err != nil {
		return nil, err
	}

	return []prometheus.Metric{
		prometheus.MustNewConstMetric(n.recordedTitleTotal, prometheus.GaugeValue, float64(recordedTitleList.TotalMatches), commonLabel...),
	}, nil
}

func (n *NasneCollector) collectReserve(client *nasneclient.NasneClient, commonLabel []string) ([]prometheus.Metric, error) {
	reservedList, err := client.GetReservedList()
	if err != nil {
		return nil, err
	}

	var conflictCount float64
	for _, r := range reservedList.Item {
		if r.ConflictID != 0 {
			conflictCount++
		}
	}

	return []prometheus.Metric{
		prometheus.MustNewConstMetric(n.reservedConflictTotal, prometheus.GaugeValue, conflictCount, commonLabel...),
	}, nil
}

func (n *NasneCollector) getCommonLabel(client *nasneclient.NasneClient) ([]string, error) {
	bn, err := client.GetBoxName()
	if err != nil {
		return nil, err
	}

	return []string{
		bn.Name,
	}, nil
}

func (n *NasneCollector) runCollect() {
	glog.V(2).Info("start collect")
	start := time.Now()

	metrics := []prometheus.Metric{}

	for _, ip := range n.nasneAddrs {
		glog.V(2).Infof("start colllect: ipaddr = %v", ip)
		startEach := time.Now()

		client, err := nasneclient.NewNasneClient(ip)
		if err != nil {
			glog.Error(err)
			continue
		}

		commonLabel, err := n.getCommonLabel(client)
		if err != nil {
			glog.Error(err)
			continue
		}

		if m, err := n.collectVersion(client, commonLabel); err != nil {
			glog.Error(err)
		} else {
			metrics = append(metrics, m...)
		}

		if m, err := n.collectHDD(client, commonLabel); err != nil {
			glog.Error(err)
		} else {
			metrics = append(metrics, m...)
		}

		if m, err := n.collectDTCPClient(client, commonLabel); err != nil {
			glog.Error(err)
		} else {
			metrics = append(metrics, m...)
		}

		if m, err := n.collectRecord(client, commonLabel); err != nil {
			glog.Error(err)
		} else {
			metrics = append(metrics, m...)
		}

		if m, err := n.collectReserve(client, commonLabel); err != nil {
			glog.Error(err)
		} else {
			metrics = append(metrics, m...)
		}

		if m, err := n.collectCollectionDuration(startEach, time.Now(), commonLabel); err != nil {
			glog.Error(err)
		} else {
			metrics = append(metrics, m...)
		}

		glog.V(2).Infof("end colllect: ipaddr = %v", ip)
	}

	if m, err := n.collectNasneCollector(start, time.Now()); err != nil {
		glog.Error(err)
	} else {
		metrics = append(metrics, m...)
	}

	n.cache.Set(metrics)

	glog.V(2).Info("end collect")
}
