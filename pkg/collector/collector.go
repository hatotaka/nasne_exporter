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
		hddSizeBytesGauge: prometheus.NewGaugeVec(
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
		hddUsageBytesGauge: prometheus.NewGaugeVec(
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
		dtcpipClientsGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "dtcpip_clients",
				Help:      "Number of clients connected with DTCP-IP.",
			},
			[]string{
				labelName,
			},
		),
		recordingsGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "recordings",
				Help:      "Number of recordings.",
			},
			[]string{
				labelName,
			},
		),
		recordedTitlesGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "recorded_titles",
				Help:      "Number of recorded titles.",
			},
			[]string{
				labelName,
			},
		),
		reservedTitlesGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "reserved_titles",
				Help:      "Number of reserved titles.",
			},
			[]string{
				labelName,
			},
		),
		reservedConflictTitlesGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "reserved_conflict_titles",
				Help:      "Number of Conflicting titles.",
			},
			[]string{
				labelName,
			},
		),
		reservedNotFoundTitlesGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "reserved_notfound_titles",
				Help:      "Number of titles that could not be found.",
			},
			[]string{
				labelName,
			},
		),
		lastCollectTileGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "last_collect_time",
				Help:      "Time of last collect metrics of nasne.",
			},
			[]string{},
		),
		collectDurationSecondsHistogram: prometheus.NewHistogramVec(
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

	infoGauge                       *prometheus.GaugeVec
	hddSizeBytesGauge               *prometheus.GaugeVec
	hddUsageBytesGauge              *prometheus.GaugeVec
	dtcpipClientsGauge              *prometheus.GaugeVec
	recordingsGauge                 *prometheus.GaugeVec
	recordedTitlesGauge             *prometheus.GaugeVec
	reservedTitlesGauge             *prometheus.GaugeVec
	reservedConflictTitlesGauge     *prometheus.GaugeVec
	reservedNotFoundTitlesGauge     *prometheus.GaugeVec
	lastCollectTileGauge            *prometheus.GaugeVec
	collectDurationSecondsHistogram *prometheus.HistogramVec
}

func (n *NasneCollector) RegisterCollectors(r *prometheus.Registry) {
	r.MustRegister(
		n.infoGauge,
		n.hddSizeBytesGauge,
		n.hddUsageBytesGauge,
		n.dtcpipClientsGauge,
		n.recordingsGauge,
		n.recordedTitlesGauge,
		n.reservedTitlesGauge,
		n.reservedConflictTitlesGauge,
		n.reservedNotFoundTitlesGauge,
		n.lastCollectTileGauge,
		n.collectDurationSecondsHistogram,
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
	n.collectDurationSecondsHistogram.With(commonLabel).Observe(end.Sub(start).Seconds())
	return nil
}

func (n *NasneCollector) collectInfo(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
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

		n.hddSizeBytesGauge.With(mergeLabels(commonLabel, labels)).Set(hddInfo.HDD.TotalVolumeSize)
		n.hddUsageBytesGauge.With(mergeLabels(commonLabel, labels)).Set(hddInfo.HDD.UsedVolumeSize)
	}

	return nil
}

func (n *NasneCollector) collectDTCPClient(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	dtcpipClientList, err := client.GetDTCPIPClientList()
	if err != nil {
		return err
	}

	n.dtcpipClientsGauge.With(commonLabel).Set(float64(dtcpipClientList.Number))

	return nil
}

func (n *NasneCollector) collectRecordings(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	boxStatusList, err := client.GetBoxStatusList()
	if err != nil {
		return err
	}

	var recordTotal float64
	if boxStatusList.TuningStatus.Status == 3 {
		recordTotal = 1
	}

	n.recordingsGauge.With(commonLabel).Set(recordTotal)

	return nil
}

func (n *NasneCollector) collectRecorded(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	recordedTitleList, err := client.GetRecordedTitleList()
	if err != nil {
		return err
	}

	n.recordedTitlesGauge.With(commonLabel).Set(float64(recordedTitleList.TotalMatches))

	return nil
}

func (n *NasneCollector) collectReserved(client *nasneclient.NasneClient, commonLabel prometheus.Labels) error {
	reservedList, err := client.GetReservedList()
	if err != nil {
		return err
	}

	var conflictCount float64
	var notFoundCount float64
	for _, r := range reservedList.Item {
		if r.EventID == nasneclient.EventIDNotFound {
			notFoundCount++
			continue
		}
		if r.ConflictID == nasneclient.ConflictIDConflictNG {
			conflictCount++
			continue
		}
	}

	n.reservedConflictTitlesGauge.With(commonLabel).Set(conflictCount)
	n.reservedNotFoundTitlesGauge.With(commonLabel).Set(notFoundCount)
	n.reservedTitlesGauge.With(commonLabel).Set(float64(reservedList.TotalMatches))

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

		if err := n.collectInfo(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectHDD(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectDTCPClient(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectRecordings(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectRecorded(client, commonLabel); err != nil {
			glog.Error(err)
		}

		if err := n.collectReserved(client, commonLabel); err != nil {
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
