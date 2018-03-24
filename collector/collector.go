package collector

import (
	"log"
	"strconv"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne-exporter/pkg/nasneclient"
	"github.com/prometheus/client_golang/prometheus"
)

func NewNasneCollector(nasneAddrs []string) prometheus.Collector {
	return &nasneCollector{
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
	}
}

type nasneCollector struct {
	nasneAddrs []string

	info                  *prometheus.Desc
	hddTotal              *prometheus.Desc
	hddUsed               *prometheus.Desc
	dtcpipClientTotal     *prometheus.Desc
	recordedTitleTotal    *prometheus.Desc
	reservedConflictTotal *prometheus.Desc
}

func (n *nasneCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- n.info
	ch <- n.hddTotal
	ch <- n.hddUsed
	ch <- n.dtcpipClientTotal
	ch <- n.recordedTitleTotal
	ch <- n.reservedConflictTotal
}

func (n *nasneCollector) Collect(ch chan<- prometheus.Metric) {
	glog.V(2).Info("start collect")

	for _, ip := range n.nasneAddrs {
		glog.V(2).Infof("start colllect: ipaddr = %v", ip)

		client, err := nasneclient.NewNasneClient(ip)
		if err != nil {
			glog.Fatal(err)
		}

		bn, err := client.GetBoxName()
		if err != nil {
			glog.Fatal(err)
		}

		{
			softwareVersion, err := client.GetSoftwareVersion()
			if err != nil {
				glog.Fatal(err)
			}

			hardwareVersion, err := client.GetHardwareVersion()
			if err != nil {
				log.Fatal(err)
			}

			ch <- prometheus.MustNewConstMetric(n.info, prometheus.GaugeValue, float64(1),
				bn.Name,
				softwareVersion.SoftwareVersion,
				strconv.Itoa(hardwareVersion.HardwareVersion),
				hardwareVersion.ProductName,
			)
		}

		{
			hddList, err := client.GetHDDList()
			if err != nil {
				glog.Fatal(err)
			}

			for _, hdd := range hddList.HDD {
				hddInfo, err := client.GetHDDInfo(hdd.ID)
				if err != nil {
					glog.Fatal(err)
				}

				labelValues := []string{
					bn.Name,
					strconv.Itoa(hddInfo.HDD.ID),
					hddInfo.HDD.Format,
					hddInfo.HDD.Name,
					hddInfo.HDD.VendorID,
					hddInfo.HDD.ProductID,
				}

				ch <- prometheus.MustNewConstMetric(n.hddTotal, prometheus.GaugeValue, hddInfo.HDD.TotalVolumeSize, labelValues...)
				ch <- prometheus.MustNewConstMetric(n.hddUsed, prometheus.GaugeValue, hddInfo.HDD.UsedVolumeSize, labelValues...)

			}
		}

		{
			dtcpipClientList, err := client.GetDTCPIPClientList()
			if err != nil {
				log.Fatal(err)
			}

			labelValues := []string{
				bn.Name,
			}

			ch <- prometheus.MustNewConstMetric(n.dtcpipClientTotal, prometheus.GaugeValue, float64(dtcpipClientList.Number), labelValues...)
		}
		{
			recordedTitleList, err := client.GetRecordedTitleList()
			if err != nil {
				log.Fatal(err)
			}

			labelValues := []string{
				bn.Name,
			}

			ch <- prometheus.MustNewConstMetric(n.recordedTitleTotal, prometheus.GaugeValue, float64(recordedTitleList.TotalMatches), labelValues...)
		}
		{
			reservedList, err := client.GetReservedList()
			if err != nil {
				log.Fatal(err)
			}

			labelValues := []string{
				bn.Name,
			}

			var conflictCount float64
			for _, r := range reservedList.Item {
				if r.ConflictID != 0 {
					conflictCount++
					glog.Info(r)
				}
			}

			ch <- prometheus.MustNewConstMetric(n.reservedConflictTotal, prometheus.GaugeValue, conflictCount, labelValues...)
		}

		glog.V(2).Infof("end colllect: ipaddr = %v", ip)
	}

	glog.V(2).Info("end collect")

}
