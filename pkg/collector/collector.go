package collector

import (
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne-exporter/pkg/nasneclient"
	"github.com/prometheus/client_golang/prometheus"
)

func NewNasneCollector(nasneAddrs []string) *NasneCollector {
	return &NasneCollector{
		nasneAddrs: nasneAddrs,

		infoGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nasne_info",
				Help: "info of nasne",
			},
			[]string{
				"name",
				"software_version",
				"hardware_version",
				"product_name",
			},
		),
		hddTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nasne_hdd_byte_total",
				Help: "nasne hdd byte total",
			},
			[]string{
				"name",
				"id",
				"format",
				"hdd_name",
				"vendor_id",
				"product_id",
			},
		),

		hddUsedGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nasne_hdd_byte_used",
				Help: "nasne hdd byte used",
			},
			[]string{
				"name",
				"id",
				"format",
				"hdd_name",
				"vendor_id",
				"product_id",
			},
		),

		dtcpipClientTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nasne_dtcpip_client_total",
				Help: "nasne dtcpip client total",
			},
			[]string{
				"name",
			},
		),

		recordedTitleTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nasne_recorded_title_total",
				Help: "number of dtcpip client",
			},
			[]string{
				"name",
			},
		),

		reservedConflictTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nasne_conflict_total",
				Help: "number of conflict",
			},
			[]string{
				"name",
			},
		),

		collectTimeGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nasne_last_collect_time",
				Help: "time of last collect",
			},
			[]string{},
		),

		totalCollectionDurationsHistogram: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "nasne_total_collection_durations_histogram_seconds",
				Help:    "Total collection latency distributions.",
				Buckets: prometheus.LinearBuckets(1, 1, 10),
			},
		),

		collectionDurationsHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "nasne_collection_durations_histogram_seconds",
				Help:    "Collection latency distributions.",
				Buckets: prometheus.LinearBuckets(1, 1, 10),
			},
			[]string{"name"},
		),
	}
}

type NasneCollector struct {
	nasneAddrs []string

	infoGauge                         *prometheus.GaugeVec
	hddTotalGauge                     *prometheus.GaugeVec
	hddUsedGauge                      *prometheus.GaugeVec
	dtcpipClientTotalGauge            *prometheus.GaugeVec
	recordedTitleTotalGauge           *prometheus.GaugeVec
	reservedConflictTotalGauge        *prometheus.GaugeVec
	collectTimeGauge                  *prometheus.GaugeVec
	totalCollectionDurationsHistogram prometheus.Histogram
	collectionDurationsHistogram      *prometheus.HistogramVec
}

func (n *NasneCollector) RegisterCollector(r *prometheus.Registry) {
	r.MustRegister(
		n.infoGauge,
		n.hddTotalGauge,
		n.hddUsedGauge,
		n.dtcpipClientTotalGauge,
		n.recordedTitleTotalGauge,
		n.reservedConflictTotalGauge,
		n.collectTimeGauge,
		n.totalCollectionDurationsHistogram,
		n.collectionDurationsHistogram,
	)
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

func (n *NasneCollector) collectNasneCollector(start, end time.Time) error {
	n.totalCollectionDurationsHistogram.Observe(end.Sub(start).Seconds())
	return nil
}

func (n *NasneCollector) collectCollectionDuration(start, end time.Time, commonLabel prometheus.Labels) error {
	n.collectionDurationsHistogram.With(commonLabel).Observe(end.Sub(start).Seconds())
	return nil
}

func (n *NasneCollector) collectVersion(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	softwareVersion, err := client.GetSoftwareVersion()
	if err != nil {
		return err
	}

	hardwareVersion, err := client.GetHardwareVersion()
	if err != nil {
		return err
	}

	labels := prometheus.Labels{
		"software_version": softwareVersion.SoftwareVersion,
		"hardware_version": strconv.Itoa(hardwareVersion.HardwareVersion),
		"product_name":     hardwareVersion.ProductName,
	}

	n.infoGauge.With(mergeLabels(commonLabel, labels)).Set(1)

	return nil
}

func (n *NasneCollector) collectHDD(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	hddList, err := client.GetHDDList()
	if err != nil {
		return err
	}

	for _, hdd := range hddList.HDD {
		hddInfo, err := client.GetHDDInfo(hdd.ID)
		if err != nil {
			glog.Fatal(err)
		}

		labels := prometheus.Labels{
			"id":         strconv.Itoa(hddInfo.HDD.ID),
			"format":     hddInfo.HDD.Format,
			"hdd_name":   hddInfo.HDD.Name,
			"vendor_id":  hddInfo.HDD.VendorID,
			"product_id": hddInfo.HDD.ProductID,
		}

		n.hddTotalGauge.With(mergeLabels(commonLabel, labels)).Set(hddInfo.HDD.TotalVolumeSize)
		n.hddUsedGauge.With(mergeLabels(commonLabel, labels)).Set(hddInfo.HDD.UsedVolumeSize)
	}

	return nil
}

func (n *NasneCollector) collectDTCPClient(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	dtcpipClientList, err := client.GetDTCPIPClientList()
	if err != nil {
		return err
	}

	n.dtcpipClientTotalGauge.With(commonLabel).Set(float64(dtcpipClientList.Number))

	return nil
}

func (n *NasneCollector) collectRecord(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	recordedTitleList, err := client.GetRecordedTitleList()
	if err != nil {
		return err
	}

	n.recordedTitleTotalGauge.With(commonLabel).Set(float64(recordedTitleList.TotalMatches))

	return nil
}

func (n *NasneCollector) collectReserve(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	reservedList, err := client.GetReservedList()
	if err != nil {
		return err
	}

	var conflictCount float64
	for _, r := range reservedList.Item {
		if r.ConflictID != 0 {
			conflictCount++
		}
	}

	n.reservedConflictTotalGauge.With(commonLabel).Set(conflictCount)

	return nil
}

func (n *NasneCollector) getCommonLabel(client *nasneclient.NasneClient) (prometheus.Labels, error) {
	bn, err := client.GetBoxName()
	if err != nil {
		return nil, err
	}

	return prometheus.Labels{
		"name": bn.Name,
	}, nil
}

func (n *NasneCollector) runCollect() {
	glog.V(2).Info("start collect")
	start := time.Now()

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

		if err := n.collectVersion(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectHDD(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectDTCPClient(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectRecord(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectReserve(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectCollectionDuration(startEach, time.Now(), commonLabel); err != nil {
			glog.Error(err)
		}

		glog.V(2).Infof("end colllect: ipaddr = %v", ip)
	}

	if err := n.collectNasneCollector(start, time.Now()); err != nil {
		glog.Error(err)
	}

	glog.V(2).Info("end collect")
}

func mergeLabels(l1, l2 prometheus.Labels) prometheus.Labels {
	l := prometheus.Labels{}

	for k, v := range l1 {
		l[k] = v
	}
	for k, v := range l2 {
		l[k] = v
	}

	return l
}
