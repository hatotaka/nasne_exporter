package collector

import (
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne_exporter/pkg/nasneclient"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "nasne"

	labelName            = "name"
	labelID              = "id"
	labelFormat          = "format"
	labelSoftwareVersion = "software_version"
	labelHardwareVersion = "hardware_version"
	labelProductName     = "product_name"
	labelHDDName         = "hdd_name"
	labelVendorID        = "vendor_id"
	labelProductID       = "product_id"
)

func NewNasneCollector(nasneAddrs []string) *NasneCollector {
	return &NasneCollector{
		nasneAddrs: nasneAddrs,

		infoGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "info",
				Help:      "Information of nasne.",
			},
			[]string{
				labelName,
				labelSoftwareVersion,
				labelHardwareVersion,
				labelProductName,
			},
		),
		hddTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "hdd_size_bytes",
				Help:      "HDD size in bytes.",
			},
			[]string{
				labelName,
				labelID,
				labelFormat,
				labelHDDName,
				labelVendorID,
				labelProductID,
			},
		),
		hddUsedGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "hdd_usage_bytes",
				Help:      "HDD usage in bytes.",
			},
			[]string{
				labelName,
				labelID,
				labelFormat,
				labelHDDName,
				labelVendorID,
				labelProductID,
			},
		),
		dtcpipClientTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "dtcpip_clients",
				Help:      "Number of clients connected with DTCP-IP.",
			},
			[]string{
				labelName,
			},
		),
		recordTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "recordings",
				Help:      "Number of recordings.",
			},
			[]string{
				labelName,
			},
		),
		recordedTitleTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "recorded_titles",
				Help:      "Number of recorded titles.",
			},
			[]string{
				labelName,
			},
		),
		reservedTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "reserved_titles",
				Help:      "Number of reserved titles.",
			},
			[]string{
				labelName,
			},
		),
		reservedConflictTotalGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "reserved_conflict_titles",
				Help:      "Number of conflict titles.",
			},
			[]string{
				labelName,
			},
		),

		collectTimeGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "last_collect_time",
				Help:      "Time of last collect metrics of nasne.",
			},
			[]string{},
		),

		collectionDurationsHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "collect_duration_seconds",
				Help:      "Collection latency distributions.",
				Buckets:   prometheus.LinearBuckets(1, 1, 10),
			},
			[]string{
				labelName,
			},
		),
	}
}

type NasneCollector struct {
	nasneAddrs []string

	infoGauge                    *prometheus.GaugeVec
	hddTotalGauge                *prometheus.GaugeVec
	hddUsedGauge                 *prometheus.GaugeVec
	dtcpipClientTotalGauge       *prometheus.GaugeVec
	recordTotalGauge             *prometheus.GaugeVec
	recordedTitleTotalGauge      *prometheus.GaugeVec
	reservedTotalGauge           *prometheus.GaugeVec
	reservedConflictTotalGauge   *prometheus.GaugeVec
	collectTimeGauge             *prometheus.GaugeVec
	collectionDurationsHistogram *prometheus.HistogramVec
}

func (n *NasneCollector) RegisterCollectors(r *prometheus.Registry) {
	r.MustRegister(
		n.infoGauge,
		n.hddTotalGauge,
		n.hddUsedGauge,
		n.dtcpipClientTotalGauge,
		n.recordTotalGauge,
		n.recordedTitleTotalGauge,
		n.reservedTotalGauge,
		n.reservedConflictTotalGauge,
		n.collectTimeGauge,
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
		labelSoftwareVersion: softwareVersion.SoftwareVersion,
		labelHardwareVersion: strconv.Itoa(hardwareVersion.HardwareVersion),
		labelProductName:     hardwareVersion.ProductName,
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
			labelID:        strconv.Itoa(hddInfo.HDD.ID),
			labelFormat:    hddInfo.HDD.Format,
			labelHDDName:   hddInfo.HDD.Name,
			labelVendorID:  hddInfo.HDD.VendorID,
			labelProductID: hddInfo.HDD.ProductID,
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

func (n *NasneCollector) collectRecordNow(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	boxStatusList, err := client.GetBoxStatusList()
	if err != nil {
		return err
	}

	var recordTotal float64
	if boxStatusList.TuningStatus.Status == 3 {
		recordTotal = 1
	}

	n.recordTotalGauge.With(commonLabel).Set(recordTotal)

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
	n.reservedTotalGauge.With(commonLabel).Set(float64(reservedList.TotalMatches))

	return nil
}

func (n *NasneCollector) getCommonLabel(client *nasneclient.NasneClient) (prometheus.Labels, error) {
	bn, err := client.GetBoxName()
	if err != nil {
		return nil, err
	}

	return prometheus.Labels{
		labelName: bn.Name,
	}, nil
}

func (n *NasneCollector) runCollect() {
	glog.V(2).Info("start collect")

	for _, ip := range n.nasneAddrs {
		glog.V(2).Infof("start colllect: ipaddr = %v", ip)
		start := time.Now()

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

		if err := n.collectRecordNow(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectRecord(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectReserve(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectCollectionDuration(start, time.Now(), commonLabel); err != nil {
			glog.Error(err)
		}

		glog.V(2).Infof("end colllect: ipaddr = %v", ip)
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
